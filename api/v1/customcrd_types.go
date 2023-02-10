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

// ContainerSpec defines the desired state of Container
type ContainerSpec struct {
	Image string `json:"image"`
	Port  int32  `json:"port"`
}

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	// +optional
	ServiceName string `json:"serviceName,omitempty"`
	ServiceType string `json:"serviceType"`
	// +optional
	ServiceNodePort int32 `json:"servicePort,omitempty"`
}

// CustomCrdSpec defines the desired state of CustomCrd
type CustomCrdSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// DeploymentName represents the name of the deployment we will create using CustomCrd
	// +optional
	DeploymentName string `json:"deploymentName,omitempty"`
	// Replicas defines number of pods will be running in the deployment
	Replicas *int32 `json:"replicas"`
	// Container contains Image and Port
	Container ContainerSpec `json:"container"`
	// Service contains ServiceName, ServiceType, ServiceNodePort
	// +optional
	Service ServiceSpec `json:"service,omitempty"`
}

// CustomCrdStatus defines the observed state of CustomCrd
type CustomCrdStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	AvailableReplicas *int32 `json:"availableReplicas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CustomCrd is the Schema for the customcrds API
type CustomCrd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomCrdSpec   `json:"spec,omitempty"`
	Status CustomCrdStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomCrdList contains a list of CustomCrd
type CustomCrdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomCrd `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomCrd{}, &CustomCrdList{})
}
