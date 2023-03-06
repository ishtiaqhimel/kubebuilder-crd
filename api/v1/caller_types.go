/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ContainerSpec defines the desired state of container
type ContainerSpec struct {
	Image         string   `json:"image"`
	Command       []string `json:"command"`
	RestartPolicy string   `json:"restartPolicy"`
}

// CallerSpec defines the desired state of Caller
type CallerSpec struct {
	Schedule  string        `json:"schedule"`
	Container ContainerSpec `json:"container"`
}

// CallerStatus defines the observed state of Caller
type CallerStatus struct {
	// +optional
	CompletedJob *int32 `json:"completedJob,omitempty"`
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
	// +optional
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Caller is the Schema for the callers API
type Caller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CallerSpec   `json:"spec,omitempty"`
	Status CallerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CallerList contains a list of Caller
type CallerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Caller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Caller{}, &CallerList{})
}
