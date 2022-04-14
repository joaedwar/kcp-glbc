package clusterworkspace

import (
	"context"
	"errors"

	"github.com/kuadrant/kcp-glbc/pkg/util/metadata"

	"github.com/kcp-dev/apimachinery/pkg/logicalcluster"

	kcpapiv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"

	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	glbcWorkspaceInitializer = "initializers.kuadarant.dev/workspace-sandbox"
	glbcAPIExportName        = "glbc"
	exportResourceAnnotation = "kuadrant.dev/export.resource"
)

// this controller watches the control cluster and mirrors cert secrets into the KCP cluster
func (c *Controller) reconcile(ctx context.Context, clusterWorkspace *tenancyv1alpha1.ClusterWorkspace) error {

	klog.Infof("reconciling %s workspace", clusterWorkspace.Name)

	if hasInitializers(clusterWorkspace) {
		// To export resources into a workspace via the api binding, the workspace
		// needs to have a status ready therefore we are adding an annotation to flag that
		// an apibinding object needs to be created.
		metadata.AddAnnotation(clusterWorkspace, exportResourceAnnotation, "true")
		_, err := c.orgClient.Cluster(logicalcluster.From(clusterWorkspace)).
			TenancyV1alpha1().
			ClusterWorkspaces().
			Update(ctx, clusterWorkspace, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		// remove initializer
		klog.Infof("removing %s workspace initializer", clusterWorkspace.Name)
		removeInitializers(clusterWorkspace)
		_, err = c.orgClient.Cluster(logicalcluster.From(clusterWorkspace)).
			TenancyV1alpha1().
			ClusterWorkspaces().
			UpdateStatus(ctx, clusterWorkspace, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	// verify if resources need to be exported to the new workspace
	if _, ok := clusterWorkspace.Annotations[exportResourceAnnotation]; ok {

		ws, err := c.orgClient.Cluster(logicalcluster.From(clusterWorkspace)).
			TenancyV1alpha1().
			ClusterWorkspaces().
			Get(ctx, clusterWorkspace.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// verify if workspace status is ready
		if ws.Status.Phase != tenancyv1alpha1.ClusterWorkspacePhaseReady {
			return errors.New("workspace is not ready, try again")
		}

		// export the glbc crds
		klog.Infof("exporting glbc resources to %s workspace", clusterWorkspace.Name)
		if err := c.reconcileAPIExportToNewWorkspace(ctx, ws); err != nil {
			return err
		}

		// remove annotation
		metadata.RemoveAnnotation(ws, exportResourceAnnotation)
		_, err = c.orgClient.Cluster(logicalcluster.From(clusterWorkspace)).
			TenancyV1alpha1().
			ClusterWorkspaces().
			Update(ctx, ws, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) reconcileAPIExportToNewWorkspace(ctx context.Context, newWorkspace *tenancyv1alpha1.ClusterWorkspace) error {

	// get api export from the loadbalancer workspace
	glbcWorkspacePath := "root:default:kcp-glbc"
	glbcAPIExpoter, err := c.orgClient.Cluster(logicalcluster.New(glbcWorkspacePath)).
		ApisV1alpha1().
		APIExports().
		Get(ctx, glbcAPIExportName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	klog.Infof("Found apiexport %s, now will create apibinding into new workspace", glbcAPIExpoter.Name)
	glbcApiBinding := kcpapiv1alpha1.APIBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: newWorkspace.GetName(),
		},
		Spec: kcpapiv1alpha1.APIBindingSpec{
			Reference: kcpapiv1alpha1.ExportReference{
				Workspace: &kcpapiv1alpha1.WorkspaceExportReference{
					WorkspaceName: "kcp-glbc",
					ExportName:    glbcAPIExpoter.Name,
				},
			},
		},
	}
	_, err = c.orgClient.Cluster(logicalcluster.From(newWorkspace).Join(newWorkspace.Name)).ApisV1alpha1().
		APIBindings().Create(ctx, &glbcApiBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	klog.Infof("apibinding %s created into new workspace %s", glbcApiBinding.Name, newWorkspace.Name)
	return nil
}

func removeInitializers(clusterWorkspace *tenancyv1alpha1.ClusterWorkspace) {
	newInitializers := make([]tenancyv1alpha1.ClusterWorkspaceInitializer, 0, len(clusterWorkspace.Status.Initializers))
	for _, i := range clusterWorkspace.Status.Initializers {
		if i != glbcWorkspaceInitializer {
			newInitializers = append(newInitializers, i)
		}
	}
	clusterWorkspace.Status.Initializers = newInitializers
}

func hasInitializers(clusterWorkspace *tenancyv1alpha1.ClusterWorkspace) bool {
	initializers := clusterWorkspace.Status.Initializers
	for _, initializer := range initializers {
		if initializer == glbcWorkspaceInitializer {
			return true
		}
	}
	return false
}
