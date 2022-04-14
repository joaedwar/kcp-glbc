package clusterworkspace

import (
	"context"

	"github.com/kcp-dev/apimachinery/pkg/logicalcluster"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	sandboxInitializer = "initializers.kuadarant.dev/workspace-sandbox"
)

// this controller watches the control cluster and mirrors cert secrets into the KCP cluster
func (c *Controller) reconcile(ctx context.Context, clusterWorkspace *tenancyv1alpha1.ClusterWorkspace) error {

	klog.Infof("reconciling clusterworkspace %s", clusterWorkspace.Name)

	hasInitializer := false
	for _, initializer := range clusterWorkspace.Status.Initializers {
		if initializer == sandboxInitializer {
			hasInitializer = true
		}
	}

	if hasInitializer {
		// TODO do something

		// remove initializer
		clusterWorkspace.Status.Initializers = make([]tenancyv1alpha1.ClusterWorkspaceInitializer, 0)
		c.orgClient.Cluster(logicalcluster.From(clusterWorkspace)).TenancyV1alpha1().ClusterWorkspaces().UpdateStatus(ctx, clusterWorkspace, metav1.UpdateOptions{})

	}
	return nil
}
