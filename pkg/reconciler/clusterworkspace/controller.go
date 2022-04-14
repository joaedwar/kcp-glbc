package clusterworkspace

import (
	"context"
	"time"

	"github.com/kcp-dev/apimachinery/pkg/logicalcluster"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	clientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	tenancyv1alpha1lister "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const controllerName = "kcp-glbc-bootstrap"

// NewController returns a new Controller which reconciles resources at workspace levels.

func NewController(config *ControllerConfig) (*Controller, error) {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &Controller{
		queue:                 queue,
		orgClient:             config.OrgClient,
		sharedInformerFactory: config.SharedInformerFactory,
	}

	c.sharedInformerFactory.Tenancy().V1alpha1().ClusterWorkspaces().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.enqueue(obj) },
		UpdateFunc: func(_, obj interface{}) { c.enqueue(obj) },
		DeleteFunc: func(obj interface{}) { c.enqueue(obj) },
	})

	c.indexer = c.sharedInformerFactory.Tenancy().V1alpha1().ClusterWorkspaces().Informer().GetIndexer()
	c.lister = c.sharedInformerFactory.Tenancy().V1alpha1().ClusterWorkspaces().Lister()
	return c, nil
}

type ControllerConfig struct {
	OrgClient             clientset.ClusterInterface
	SharedInformerFactory informers.SharedInformerFactory
}

type Controller struct {
	queue                 workqueue.RateLimitingInterface
	sharedInformerFactory informers.SharedInformerFactory
	indexer               cache.Indexer
	lister                tenancyv1alpha1lister.ClusterWorkspaceLister
	orgClient             clientset.ClusterInterface
}

func (c *Controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.queue.AddRateLimited(key)
}

func (c *Controller) Start(ctx context.Context, numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.InfoS("Starting workers", "controller", controllerName)
	defer klog.InfoS("Stopping workers", "controller", controllerName)

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(ctx, c.startWorker, time.Second)
	}

	<-ctx.Done()
}

func (c *Controller) startWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	k, quit := c.queue.Get()
	if quit {
		return false
	}
	key := k.(string)

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(key)

	err := c.process(ctx, key)
	c.handleErr(err, key)
	return true
}

func (c *Controller) handleErr(err error, key string) {
	// Reconcile worked, nothing else to do for this workqueue item.
	if err == nil {
		c.queue.Forget(key)
		return
	}

	// Re-enqueue up to 5 times.
	num := c.queue.NumRequeues(key)
	if num < 5 {
		klog.Errorf("Error reconciling key %q, retrying... (#%d): %v", key, num, err)
		c.queue.AddRateLimited(key)
		return
	}

	// Give up and report error elsewhere.
	c.queue.Forget(key)
	runtime.HandleError(err)
	klog.Infof("Dropping key %q after failed retries: %v", key, err)
}

func (c *Controller) process(ctx context.Context, key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		klog.Infof("Object with key %q was deleted", key)
		return nil
	}
	current := obj.(*tenancyv1alpha1.ClusterWorkspace)

	previous := current.DeepCopy()
	if err := c.reconcile(ctx, current); err != nil {
		return err
	}
	// If the object being reconciled changed as a result, update it.
	if !equality.Semantic.DeepEqual(previous, current) {
		_, err := c.orgClient.Cluster(logicalcluster.From(current)).TenancyV1alpha1().ClusterWorkspaces().Update(ctx, current, metav1.UpdateOptions{})
		return err
	}
	return nil
}
