package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/imjasonh/cluster-controller/pkg/apis/cluster/v1alpha1"
	"github.com/imjasonh/cluster-controller/pkg/client/clientset/versioned"
	"github.com/imjasonh/cluster-controller/pkg/client/informers/externalversions"
	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const resyncPeriod = 10 * time.Hour

func main() {
	flag.Parse()
	ctx := context.Background()

	// Connect to kcp
	cfg := config.GetConfigOrDie()

	// Ensure kcp knows about the Cluster CRD.
	apixclient := apix.NewForConfigOrDie(cfg)
	if _, err := apixclient.CustomResourceDefinitions().Get(ctx, "clusters.cluster.example.dev", metav1.GetOptions{}); err != nil {
		log.Fatalf("Failed to get Cluster CRD: %v", err)
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

	l, err := sif.Cluster().V1alpha1().Clusters().Lister().List(labels.Everything())
	if err != nil {
		log.Fatalf("listing Clusters: %v", err)
	}
	for _, i := range l {
		log.Println("-- found cluster", i.Name)
	}
	<-stopCh
}

type reconciler struct {
}

func (r reconciler) reconcileCluster(ctx context.Context, c *v1alpha1.Cluster) {
	log.Println("saw cluster", c.Name)
}
