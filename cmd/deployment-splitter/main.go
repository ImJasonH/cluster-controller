package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	clusterinformer "github.com/imjasonh/cluster-controller/pkg/client/injection/informers/cluster/v1alpha1/cluster"
	clusterlister "github.com/imjasonh/cluster-controller/pkg/client/listers/cluster/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	deploymentreconciler "knative.dev/pkg/client/injection/kube/reconciler/apps/v1/deployment"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	kreconciler "knative.dev/pkg/reconciler"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const resyncPeriod = 10 * time.Hour

func main() {
	flag.Parse()
	ctx := context.Background()

	// Connect to kcp
	cfg := config.GetConfigOrDie()

	// Ensure kcp knows about the Cluster and Deployment CRDs.
	apixclient := apix.NewForConfigOrDie(cfg)
	for _, crd := range []string{
		"clusters.cluster.example.dev",
		"deployments.apps",
	} {
		if _, err := apixclient.CustomResourceDefinitions().Get(ctx, "clusters.cluster.example.dev", metav1.GetOptions{}); err != nil {
			log.Fatalf("Failed to get %q CRD: %v", crd, err)
		}
	}

	// TODO: react to clusters going away or becoming unhealthy by rebalancing replicas.

	// Start controller watching for Deployments.
	sharedmain.MainWithConfig(ctx, "deployment-splitter", cfg,
		func(ctx context.Context, _ configmap.Watcher) *controller.Impl {
			r := &reconciler{
				kube:             kubeclient.Get(ctx),
				clusterlister:    clusterinformer.Get(ctx).Lister(),
				deploymentlister: deploymentinformer.Get(ctx).Lister(),
			}
			return deploymentreconciler.NewImpl(ctx, r, func(impl *controller.Impl) controller.Options {
				return controller.Options{
					AgentName: "deployment-splitter",
				}
			})
		})
}

type reconciler struct {
	clusterlister    clusterlister.ClusterLister
	deploymentlister appsv1lister.DeploymentLister
	kube             kubernetes.Interface
}

func (r reconciler) ReconcileKind(ctx context.Context, d *appsv1.Deployment) kreconciler.Event {
	logger := logging.FromContext(ctx)

	// If this is a virtual deployment, get other related virtual
	// deployments and collate status back to umbrella deployment.
	if isVirtual(d) {
		sel, err := labels.Parse("owned-by == " + d.Labels["owned-by"])
		if err != nil {
			return fmt.Errorf("Failed to parse label selector: %v", err)
		}
		vds, err := r.deploymentlister.Deployments(d.Namespace).List(sel)
		if err != nil {
			return fmt.Errorf("listing other virtual clusters for %q: %v", d.Labels["owned-by"], err)
		}
		combineStatus(vds, d)
		return nil
	}

	cls, err := r.clusterlister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("listing clusters: %v", err)
	}
	if len(cls) == 0 {
		d.Status.Conditions = []appsv1.DeploymentCondition{{
			Type:    appsv1.DeploymentProgressing,
			Status:  corev1.ConditionFalse,
			Reason:  "NoRegisteredClusters",
			Message: "kcp has no clusters registered to receive Deployments",
		}}
		return nil
	}
	if len(cls) == 1 {
		// nothing to split, just label Deployment for the only cluster.
		if d.Labels == nil {
			d.Labels = map[string]string{}
		}

		// TODO: munge cluster name
		d.Labels["cluster"] = cls[0].Name
		return nil
	}

	// Get virtual deployments owned by this (real) one.
	sel, err := labels.Parse("owned-by == " + d.Name)
	if err != nil {
		return fmt.Errorf("Failed to parse label selector: %v", err)
	}
	vds, err := r.deploymentlister.Deployments(d.Namespace).List(sel)
	if err != nil {
		return fmt.Errorf("listing virtual clusters for %q: %v", d.Name, err)
	}
	if len(vds) > 0 {
		// Don't need to create virtual deployments, only updates.
		return nil
	}

	// If there are >1 Clusters, create a virtual Deployment labeled/named for each Cluster with a subset of replicas requested.
	// TODO: assign replicas unevenly based on load/scheduling.
	replicasEach := *d.Spec.Replicas / int32(len(cls))
	for _, c := range cls {
		vd := d.DeepCopy()

		// TODO: munge cluster name
		vd.Name = "virtual-deployment-" + c.Name

		if vd.Labels == nil {
			vd.Labels = map[string]string{}
		}
		vd.Labels["cluster"] = c.Name
		vd.Labels["owned-by"] = c.Name

		vd.Spec.Replicas = &replicasEach

		// Set OwnerReference so deleting the Deployment deletes all virtual deployments.
		vd.OwnerReferences = []metav1.OwnerReference{{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			UID:        d.UID,
		}}

		// TODO: munge namespace
		if _, err := r.kube.AppsV1().Deployments(d.Namespace).Create(ctx, vd, metav1.CreateOptions{}); err != nil {
			logger.Errorf("creating deployment %q for cluster %q: %v", vd.Name, c.Name, err)
		}
	}

	return nil
}

func isVirtual(d *appsv1.Deployment) bool {
	if d.Labels == nil || d.Labels["owned-by"] == "" {
		return false
	}

	for _, or := range d.OwnerReferences {
		if or.APIVersion == "apps/v1" && or.Kind == "Deployment" {
			return true
		}
	}
	return false
}

func combineStatus(vds []*appsv1.Deployment, d *appsv1.Deployment) {
	// TODO: actually merge conditions.
	d.Status.Conditions = []appsv1.DeploymentCondition{{
		Type:   appsv1.DeploymentProgressing,
		Status: corev1.ConditionTrue,
	}}

	d.Status.Replicas = 0
	d.Status.UpdatedReplicas = 0
	d.Status.ReadyReplicas = 0
	d.Status.AvailableReplicas = 0
	d.Status.UnavailableReplicas = 0
	for _, vd := range vds {
		d.Status.Replicas += vd.Status.Replicas
		d.Status.UpdatedReplicas += d.Status.UpdatedReplicas
		d.Status.ReadyReplicas += d.Status.ReadyReplicas
		d.Status.AvailableReplicas += d.Status.AvailableReplicas
		d.Status.UnavailableReplicas += d.Status.UnavailableReplicas
	}
}
