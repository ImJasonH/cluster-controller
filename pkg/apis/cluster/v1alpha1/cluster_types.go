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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// Cluster describes a member cluster.
//
// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Cluster struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state.
	// +optional
	Spec ClusterSpec `json:"spec,omitempty"`

	// Status communicates the observed state.
	// +optional
	Status ClusterStatus `json:"status,omitempty"`
}

var (
	// Check that Cluster can be validated and defaulted.
	_ apis.Validatable   = (*Cluster)(nil)
	_ apis.Defaultable   = (*Cluster)(nil)
	_ kmeta.OwnerRefable = (*Cluster)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*Cluster)(nil)
)

// ClusterSpec holds the desired state of the Cluster (from the client).
type ClusterSpec struct {
	// ServiceName holds the name of the Kubernetes Service to expose as an "addressable".
	ServiceName string `json:"serviceName"`
}

const (
	// ClusterConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	ClusterConditionReady = apis.ConditionReady
)

// ClusterStatus communicates the observed state of the Cluster (from the controller).
type ClusterStatus struct {
	duckv1.Status `json:",inline"`
}

// ClusterList is a list of Cluster resources
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Cluster `json:"items"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (as *Cluster) GetStatus() *duckv1.Status {
	return &as.Status.Status
}
