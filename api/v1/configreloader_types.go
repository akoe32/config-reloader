/*
Copyright 2025.

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

// ConfigReloaderSpec defines the desired state of ConfigReloader.
type ConfigReloaderSpec struct {
	WorkloadName  string `json:"workloadName"`
	ConfigmapName string `json:"configmapName,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
	ResourceType  string `json:"resourceType,omitempty"`
	WorkloadType  string `json:"workloadType,omitempty"`
}

// ConfigReloaderStatus defines the observed state of ConfigReloader.
type ConfigReloaderStatus struct {
	LastReloadTime metav1.Time `json:"lastReloadTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ConfigReloader is the Schema for the configreloaders API.
type ConfigReloader struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigReloaderSpec   `json:"spec,omitempty"`
	Status ConfigReloaderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigReloaderList contains a list of ConfigReloader.
type ConfigReloaderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigReloader `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigReloader{}, &ConfigReloaderList{})
}
