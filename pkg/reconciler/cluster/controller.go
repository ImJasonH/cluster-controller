/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"context"

	"knative.dev/pkg/tracker"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	clusterinformer "github.com/imjasonh/cluster-controller/pkg/client/injection/informers/cluster/v1alpha1/cluster"
	clusterreconciler "github.com/imjasonh/cluster-controller/pkg/client/injection/reconciler/cluster/v1alpha1/cluster"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	clusterInformer := clusterinformer.Get(ctx)

	r := &Reconciler{}
	impl := clusterreconciler.NewImpl(ctx, r)
	r.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))

	logger.Info("Setting up event handlers.")

	clusterInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
