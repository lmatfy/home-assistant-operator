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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstanceHost describes a single host pointing to a Home Assistant service.
type InstanceHost struct {
	// host is the host which points to the Home Assistant Service.
	Host string `json:"host,omitempty"`

	// path is the path used for Ingress. It defaults to "/".
	//+optional
	Path string `json:"path,omitempty"`

	// servicePort is the port of the Home Assistant service. It defaults to 8123.
	//+optional
	ServicePort int32 `json:"servicePort,omitempty"`

	// secretName of an existing Secret containing the TLS certificate. It defaults to
	// a generated name containing the name of the Instance resource.
	//+optional
	SecretName string `json:"secretName,omitempty"`

	// ingressClassName is the name of an IngressClass cluster resource. It defaults to
	// the clusters default IngressClass class.
	//+optional
	IngressClassName *string `json:"ingressClassName,omitempty"`
}

// InstanceServicePort describes a ports of the for Home Assistant service.
type InstanceServicePort struct {
	// servicePort is the port of the Home Assistant service.
	ServicePort int32 `json:"servicePort,omitempty"`

	// podPort is the port of the Home Assistant pod.
	PodPort int32 `json:"podPort,omitempty"`

	// The port on each node on which this service is exposed when type is
	// NodePort or LoadBalancer.  Usually assigned by the system. If a value is
	// specified, in-range, and not in use it will be used, otherwise the
	// operation will fail.  If not specified, a port will be allocated if this
	// Service requires one.  If this field is specified when creating a
	// Service which does not need it, creation will fail. This field will be
	// wiped when updating a Service to no longer need it (e.g. changing type
	// from NodePort to ClusterIP).
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	// +optional
	NodePort int32 `json:"nodePort,omitempty"`
}

// Service Type string describes ingress methods for a Home Assistant service.
// +enum
type ServiceType string

const (
	// ServiceTypeClusterIP means a service will only be accessible inside the
	// cluster, via the cluster IP.
	ServiceTypeClusterIP ServiceType = "ClusterIP"

	// ServiceTypeNodePort means a service will be exposed on one port of
	// every node, in addition to 'ClusterIP' type.
	ServiceTypeNodePort ServiceType = "NodePort"

	// ServiceTypeLoadBalancer means a service will be exposed via an
	// external load balancer (if the cloud provider supports it), in addition
	// to 'NodePort' type.
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
)

// InstanceHost describes the service for Home Assistant.
type InstanceService struct {
	// type determines how the Service is exposed. Defaults to ClusterIP. Valid
	// options are ClusterIP, NodePort, and LoadBalancer.
	// "ClusterIP" allocates a cluster-internal IP address for load-balancing
	// to endpoints. Endpoints are determined by the selector or if that is not
	// specified, by manual construction of an Endpoints object or
	// EndpointSlice objects. If clusterIP is "None", no virtual IP is
	// allocated and the endpoints are published as a set of endpoints rather
	// than a virtual IP.
	// "NodePort" builds on ClusterIP and allocates a port on every node which
	// routes to the same endpoints as the clusterIP.
	// "LoadBalancer" builds on NodePort and creates an external load-balancer
	// (if supported in the current cloud) which routes to the same endpoints
	// as the clusterIP.
	// +optional
	Type ServiceType `json:"type,omitempty"`

	// port is a list of service and pod ports.
	Ports []InstanceServicePort `json:"ports,omitempty"`
}

// InstanceSpec provides the specification of an Instance.
type InstanceSpec struct {
	// version is the actual Home Assistant version to be installed.
	Version string `json:"version,omitempty"`

	// hosts is a list of hosts containing Ingress and Secret information.
	//+optional
	Hosts []InstanceHost `json:"host,omitempty"`

	// service is the Home Assistant service.
	//+optional
	Service InstanceService `json:"service,omitempty"`
}

// InstanceStatus defines the observed state of Instance.
type InstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
