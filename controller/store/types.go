// Copyright 2019 HAProxy Technologies LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

// ServicePort describes port of a service
type ServicePort struct {
	Name     string
	Protocol string
	Port     int64
	Status   Status
}

type HAProxySrv struct {
	// Srv disabled is srv with address ""
	Name     string
	Address  string
	Modified bool
	Port     int64
}

// PortEndpoints describes endpoints of a service port
type PortEndpoints struct {
	Port int64
	//BackendName     string // For runtime operations
	DynUpdateFailed bool
	AddrCount       int
	AddrNew         map[string]struct{}
}

// Endpoints describes endpoints of a service
type Endpoints struct {
	SliceName string
	Namespace string
	Service   string
	Ports     map[string]*PortEndpoints // Ports[portName]
	Status    Status
}

// Service is useful data from k8s structures about service
type Service struct {
	Namespace   string
	Name        string
	Ports       []ServicePort
	Addresses   []string // Used only for publish-service
	DNS         string
	Annotations map[string]string
	Status      Status
}

type Address struct {
	Address string
	Port    int64
}

// Namespace is useful data from k8s structures about namespace
type Namespace struct {
	_         [0]int
	Name      string
	Relevant  bool
	Ingresses map[string]*Ingress
	Endpoints map[string]map[string]*Endpoints // <- comes from endpoints, but slices nngh
	Services  map[string]*Service
	Secret    map[string]*Secret

	// we can't have individual slice based HAProxySrvs. Why? It must include all items for the syncing. Otherwise it's not possible to know what to disable.
	HAProxySrvs map[string]map[string][]*HAProxySrv // service :: port :: slice of haproxysrv
	BackendName map[string]string                   // For runtime operations, goes together with HAProxySrvs

	// we can't have individual slice based NewAddresses. Why? Same as HAProxySrvs. We don't know what to items have been removed. The sync must have the full state of all slices available.
	NewAddresses map[string]map[string]map[string]*Address // service :: port :: set of addr

	Status Status
}

type IngressClass struct {
	APIVersion string
	Name       string
	Controller string
	Status     Status
}

// IngressPath is useful data from k8s structures about ingress path
type IngressPath struct {
	SvcName          string
	SvcPortInt       int64
	SvcPortString    string
	SvcPortResolved  *ServicePort
	Path             string
	PathTypeMatch    string
	IsDefaultBackend bool
	Status           Status
}

// IngressRule is useful data from k8s structures about ingress rule
type IngressRule struct {
	Host   string
	Paths  map[string]*IngressPath
	Status Status
}

// Ingress is useful data from k8s structures about ingress
type Ingress struct {
	// Required for K8s.UpdateIngressStatus to select proper versioned Client Set
	APIVersion     string
	Namespace      string
	Name           string
	Class          string
	Annotations    map[string]string
	Rules          map[string]*IngressRule
	DefaultBackend *IngressPath
	TLS            map[string]*IngressTLS
	Status         Status
}

// IngressTLS describes the transport layer security associated with an Ingress.
type IngressTLS struct {
	Host       string
	SecretName string
	Status     Status
}

type ConfigMaps struct {
	Main         *ConfigMap
	TCPServices  *ConfigMap
	Errorfiles   *ConfigMap
	PatternFiles *ConfigMap
}

// ConfigMap is useful data from k8s structures about configmap
type ConfigMap struct {
	Namespace   string
	Name        string
	Loaded      bool
	Annotations map[string]string
	Status      Status
}

// Secret is useful data from k8s structures about secret
type Secret struct {
	Namespace string
	Name      string
	Data      map[string][]byte
	Status    Status
}
