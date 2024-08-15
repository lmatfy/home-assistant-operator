/*
Copyright 2024.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstancePhase is a label for the condition of an instance at the current time.
// +enum
type InstancePhase string

// These are the valid statuses of instances.
const (
	// InstancePending means the instance has been accepted by the system, but one or more of the resources
	// has not been created.
	InstancePending InstancePhase = "Pending"
	// InstanceCreated means all instance resources has been created.
	InstanceCreated InstancePhase = "Created"
	// InstanceFailed means that at least one resource failed to be created.
	InstanceFailed InstancePhase = "Failed"
)

// Persistence describes the volume for Home Assistant.
type Persistence struct {
	// Size specifies the size of the PersistentVolumeClaim. Defaults to 1Gbi
	// +optional
	Size string `json:"size,omitempty"`
	// StorageClassName is the name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// Ingress describes the Ingress resource for Home Assistant.
type Ingress struct {
	// Enabled determines wheater an Ingress resource must be created. Defaults to false.
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Host specifies the host of the Ingress resource.
	// +optional
	Host string `json:"host,omitempty"`
	// SecretName specifies the name of the secret containing the TLS certificate.
	// +optional
	SecretName string `json:"secretName,omitempty"`
	// ingressClassName is the name of an IngressClass cluster resource. Ingress
	// controller implementations use this field to know whether they should be
	// serving this Ingress resource, by a transitive connection
	// (controller -> IngressClass -> Ingress resource). Although the
	// `kubernetes.io/ingress.class` annotation (simple constant name) was never
	// formally defined, it was widely supported by Ingress controllers to create
	// a direct binding between Ingress controller and Ingress resources. Newly
	// created Ingress resources should prefer using the field. However, even
	// though the annotation is officially deprecated, for backwards compatibility
	// reasons, ingress controllers should still honor that annotation if present.
	// +optional
	IngressClassName *string `json:"ingressClassName,omitempty"`
}

// InstanceSpec provides the specification of a Home Assistant instance.
type InstanceSpec struct {
	// Labels is a list of additional labels which should be added.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations is a list of additional annotations which should be added.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Version is the actual Home Assistant version to be installed.
	Version string `json:"version,omitempty"`
	// List of environment variables to set.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// Host networking requested. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// Default to false.
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity corev1.Affinity `json:"affinity,omitempty"`

	// Ingress describes the Ingress for Home Assistant.
	// +optional
	Ingress Ingress `json:"ingress,omitempty"`
	// Persistence defines the desired characteristics of the Home Assistant volume.
	// +optional
	Persistence Persistence `json:"persistence,omitempty"`
}

// InstanceStatus defines the observed state of Instance.
type InstanceStatus struct {
	// The phase of an Instance is a simple, high-level summary of the Instance resources.
	// +optional
	Phase InstancePhase `json:"phase,omitempty"`
	// A human readable message indicating details about why the pod is in this condition.
	// +optional
	Message string `json:"message,omitempty"`
	// A brief CamelCase message indicating details about why the instance is in this state.
	// +optional
	Reason string `json:"reason,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
