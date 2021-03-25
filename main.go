package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/imjasonh/cluster-controller/pkg/apis/cluster/v1alpha1"
	"github.com/imjasonh/cluster-controller/pkg/client/clientset/versioned"
	"github.com/imjasonh/cluster-controller/pkg/client/informers/externalversions"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const resyncPeriod = 10 * time.Hour

func main() {
	flag.Parse()
	ctx := context.Background()

	// Connect to KCP
	cfg := config.GetConfigOrDie()

	// Ensure KCP knows about the Cluster CRD.
	apixclient := apix.NewForConfigOrDie(cfg)
	trueBool := true
	// TODO: This should just apply the YAML...
	if _, err := apixclient.CustomResourceDefinitions().Create(ctx, &apixv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusters.cluster.example.dev",
		},
		Spec: apixv1.CustomResourceDefinitionSpec{
			Group: "cluster.example.dev",
			Scope: apixv1.ClusterScoped,
			Names: apixv1.CustomResourceDefinitionNames{
				Kind:       "Cluster",
				Plural:     "clusters",
				Singular:   "cluster",
				Categories: []string{"all"},
			},
			Versions: []apixv1.CustomResourceDefinitionVersion{{
				Name:    "v1alpha1",
				Served:  true,
				Storage: true,
				Subresources: &apixv1.CustomResourceSubresources{
					Status: &apixv1.CustomResourceSubresourceStatus{},
				},
				// TODO: do better than this.
				Schema: &apixv1.CustomResourceValidation{
					OpenAPIV3Schema: &apixv1.JSONSchemaProps{
						Type:                   "object",
						XPreserveUnknownFields: &trueBool,
					},
				},
			}},
		},
	}, metav1.CreateOptions{}); k8serrors.IsAlreadyExists(err) {
		log.Println("KCP already has Cluster CRD!")
	} else if err != nil {
		log.Fatalf("KCP rejected Cluster CRD Create: %v", err)
	}

	// Watch for Cluster resources.
	client := versioned.NewForConfigOrDie(cfg)
	r := reconciler{}
	sif := externalversions.NewSharedInformerFactory(client, resyncPeriod)
	stopCh := make(chan struct{})
	sif.Cluster().V1alpha1().Clusters().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { r.reconcileCluster(ctx, obj.(*v1alpha1.Cluster)) },
		UpdateFunc: func(_, obj interface{}) { r.reconcileCluster(ctx, obj.(*v1alpha1.Cluster)) },
		DeleteFunc: func(obj interface{}) { r.reconcileCluster(ctx, obj.(*v1alpha1.Cluster)) },
	})
	sif.WaitForCacheSync(stopCh)
	sif.Start(stopCh)
	<-stopCh
}

type reconciler struct {
}

func (r reconciler) reconcileCluster(ctx context.Context, c *v1alpha1.Cluster) {
	log.Println("saw cluster", c.Name)
}
