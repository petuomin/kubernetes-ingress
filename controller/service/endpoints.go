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

package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-test/deep"

	"github.com/haproxytech/client-native/v2/models"

	"github.com/haproxytech/kubernetes-ingress/controller/annotations"
	"github.com/haproxytech/kubernetes-ingress/controller/haproxy"
	"github.com/haproxytech/kubernetes-ingress/controller/haproxy/api"
	"github.com/haproxytech/kubernetes-ingress/controller/store"
)

// HandleEndpoints lookups the IngressPath related endpoints and handles corresponding backend servers configuration in HAProxy
func (s *SvcContext) HandleEndpoints(client api.HAProxyClient, k8sStore store.K8s, certs *haproxy.Certificates) (reload bool) {
	var srvsScaled, srvsActiveAnn bool
	var srv, oldSrv *models.Server
	/*endpoints, err := s.getEndpoints(store)
	if err != nil {
		logger.Warningf("Ingress '%s/%s': %s", s.ingress.Namespace, s.ingress.Name, err)
		return
	}
	*/

	ns := k8sStore.Namespaces[s.service.Namespace]
	if ns == nil {
		logger.Warningf("Ingress '%s/%s': Not found", s.ingress.Namespace, s.ingress.Name)
		return
	}
	sp := s.path.SvcPortResolved

	// set backendName in store for runtime updates.
	ns.HAProxyConfig[s.service.Name].BackendName[sp.Name] = s.backendName

	newAddresses := ns.HAProxyConfig[s.service.Name].NewAddresses[sp.Name]

	if ns.HAProxyConfig[s.service.Name].HAProxySrvs[sp.Name] == nil {
		tmp := make([]*store.HAProxySrv, 0, len(newAddresses))
		ns.HAProxyConfig[s.service.Name].HAProxySrvs[sp.Name] = &tmp
	}
	HAProxySrvs := ns.HAProxyConfig[s.service.Name].HAProxySrvs[sp.Name]

	if s.service.DNS == "" {
		srvsScaled = s.scaleHAProxySrvs(&newAddresses, HAProxySrvs, k8sStore)
	}
	srv = &models.Server{}
	annotations.HandleServerAnnotations(
		srv,
		k8sStore,
		client,
		certs,
		s.service.Annotations,
		s.ingress.Annotations,
		s.store.ConfigMaps.Main.Annotations,
	)
	if !s.newBackend {
		oldSrv, _ = client.ServerGet("SRV_1", s.backendName)
		srv.Name = "SRV_1"
		result := deep.Equal(oldSrv, srv)
		if len(result) != 0 {
			srvsActiveAnn = true
			logger.Debugf("Ingress '%s/%s': server options for backend '%s' were updated:%s\nReload required", s.ingress.Namespace, s.ingress.Name, s.backendName, result)
		}
	}
	for _, srvSlot := range *HAProxySrvs {
		if srvSlot.Modified || s.newBackend || srvsActiveAnn {
			logger.Debugf("Name: %s  Address: %s  Port: %s", srvSlot.Name, srvSlot.Address, srvSlot.Port)

			s.updateHAProxySrv(client, *srv, *srvSlot, srvSlot.Port)
		}
	}

	return srvsScaled || srvsActiveAnn
}

// updateHAProxySrv updates corresponding HAProxy backend server or creates one if it does not exist
func (s *SvcContext) updateHAProxySrv(client api.HAProxyClient, srv models.Server, srvSlot store.HAProxySrv, port int64) {
	srv.Name = srvSlot.Name
	srv.Port = &port
	// Enabled/Disabled
	if srvSlot.Address == "" {
		srv.Address = "127.0.0.1"
		srv.Maintenance = "enabled"
	} else {
		srv.Address = srvSlot.Address
		srv.Maintenance = "disabled"
	}
	// Update server
	errAPI := client.BackendServerEdit(s.backendName, srv)
	if errAPI == nil {
		logger.Tracef("Updating server '%s/%s' %s", s.backendName, srv.Name, srv.Port)
		return
	}
	// Create server
	if strings.Contains(errAPI.Error(), "does not exist") {
		logger.Tracef("Creating server '%s/%s'", s.backendName, srv.Name)
		logger.Error(client.BackendServerCreate(s.backendName, srv))
	}
}

// scaleHAproxySrvs adds servers to match available addresses
func (s *SvcContext) scaleHAProxySrvs(newAddresses *map[string]*store.Address, HAProxySrvs *[]*store.HAProxySrv, k8sStore store.K8s) (reload bool) {
	var flag bool
	var srvSlots int
	var disabled []*store.HAProxySrv
	// Add disabled HAProxySrvs to match "scale-server-slots"
	// scale-server-slots has a default value in defaultAnnotations
	// "servers-increment", "server-slots" are legacy annotations
	for _, annotation := range []string{"servers-increment", "server-slots", "scale-server-slots"} {
		annServerSlots := k8sStore.GetValueFromAnnotations(annotation, k8sStore.ConfigMaps.Main.Annotations)
		if annServerSlots != "" {
			if value, err := strconv.Atoi(annServerSlots); err == nil {
				srvSlots = value
				break
			} else {
				logger.Error(err)
			}
		}
	}

	for len(*HAProxySrvs) < srvSlots {
		srv := &store.HAProxySrv{
			Name:     fmt.Sprintf("SRV_%d", len(*HAProxySrvs)+1),
			Address:  "",
			Modified: true,
		}
		*HAProxySrvs = append(*HAProxySrvs, srv)
		disabled = append(disabled, srv)
		flag = true
	}
	if flag {
		reload = true
		logger.Debugf("Server slots in backend '%s' scaled to match scale-server-slots value: %d, reload required", s.backendName, srvSlots)
	}
	// Configure remaining addresses in available HAProxySrvs
	flag = false
	for addr, Address := range *newAddresses {
		if len(disabled) != 0 {
			disabled[0].Address = addr
			disabled[0].Modified = true
			disabled[0].Port = Address.Port
			disabled = disabled[1:]
		} else {
			srv := &store.HAProxySrv{
				Name:     fmt.Sprintf("SRV_%d", len(*HAProxySrvs)+1),
				Address:  addr,
				Modified: true,
				Port:     Address.Port,
			}
			*HAProxySrvs = append(*HAProxySrvs, srv)
			flag = true
		}
		delete(*newAddresses, addr)
	}
	if flag {
		reload = true
		logger.Debugf("Server slots in backend '%s' scaled to match available endpoints, reload required", s.backendName)
	}
	return reload
}

/*
func (s *SvcContext) getEndpoints(k8s store.K8s) (endpoints *store.Endpoints, err error) {
	var ok bool
	var e map[string]*store.Endpoints
	if ns := k8s.Namespaces[s.service.Namespace]; ns != nil {
		e, ok = ns.Endpoints[s.service.Name]
	}
	if !ok {
		if s.service.DNS != "" {
			return nil, fmt.Errorf("skipped TODO") //s.getExternalNameEndpoints()
		}
		return nil, fmt.Errorf("no Endpoints for service '%s'", s.service.Name)
	}
	sp := s.path.SvcPortResolved
	if sp != nil {
		for sliceName := range e {
			for portName, sliceEndpoints := range slice {
				if portName == sp.Name || endpoints.Port == sp.Port {
					endpointsList = append(endpointsList, endpoints)
				}
			}
		}
		return endpointsList[0], nil
	}
	if s.path.SvcPortString != "" {
		return nil, fmt.Errorf("no matching endpoints for service '%s' and port '%s'", s.service.Name, s.path.SvcPortString)
	}
	return nil, fmt.Errorf("no matching endpoints for service '%s' and port '%d'", s.service.Name, s.path.SvcPortInt)
}

func (s *SvcContext) getExternalNameEndpoints() (endpoints *store.PortEndpoints, err error) {
	logger.Tracef("Configuring service '%s', of type ExternalName", s.service.Name)
	var port int64
	for _, sp := range s.service.Ports {
		if sp.Name == s.path.SvcPortString || sp.Port == s.path.SvcPortInt {
			port = sp.Port
		}
	}
	if port == 0 {
		ingressPort := s.path.SvcPortString
		if s.path.SvcPortInt != 0 {
			ingressPort = fmt.Sprintf("%d", s.path.SvcPortInt)
		}
		return nil, fmt.Errorf("service '%s': service port '%s' not found", s.service.Name, ingressPort)
	}
	endpoints = &store.PortEndpoints{
		Port: port,
		HAProxySrvs: []*store.HAProxySrv{
			{
				Name:     "SRV_1",
				Address:  s.service.DNS,
				Modified: true,
			},
		},
	}
	return endpoints, nil
}
*/
