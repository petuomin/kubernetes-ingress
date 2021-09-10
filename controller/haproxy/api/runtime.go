package api

import (
	"github.com/haproxytech/client-native/v2/models"

	"github.com/haproxytech/kubernetes-ingress/controller/store"
	"github.com/haproxytech/kubernetes-ingress/controller/utils"
)

func (c *clientNative) ExecuteRaw(command string) (result []string, err error) {
	return c.nativeAPI.Runtime.ExecuteRaw(command)
}

func (c *clientNative) SetServerAddr(backendName string, serverName string, ip string, port int) error {
	return c.nativeAPI.Runtime.SetServerAddr(backendName, serverName, ip, port)
}

func (c *clientNative) SetServerState(backendName string, serverName string, state string) error {
	return c.nativeAPI.Runtime.SetServerState(backendName, serverName, state)
}

func (c *clientNative) SetMapContent(mapFile string, payload string) error {
	err := c.nativeAPI.Runtime.ClearMap(mapFile, false)
	if err != nil {
		return err
	}
	return c.nativeAPI.Runtime.AddMapPayload(mapFile, payload)
}

func (c *clientNative) GetMap(mapFile string) (*models.Map, error) {
	return c.nativeAPI.Runtime.GetMap(mapFile)
}

// SyncBackendSrvs syncs states and addresses of a backend servers with corresponding endpoints.
func (c *clientNative) SyncBackendSrvs(BackendName string, haproxySrvs *[]*store.HAProxySrv, newAddresses map[string]*store.Address) error {
	if BackendName == "" {
		return nil
	}

	portChanged := false // newEndpoints.Port != oldEndpoints.Port
	// Disable stale entries from HAProxySrvs
	// and provide list of Disabled Srvs
	var disabled []*store.HAProxySrv
	var errors utils.Errors
	// Delete any item from AddrNew that existed already in HAProxySrvs
	for i, srv := range *haproxySrvs {
		srv.Modified = portChanged || srv.Modified
		if _, ok := newAddresses[srv.Address]; ok {
			delete(newAddresses, srv.Address)
		} else {
			// if entry in HAProxySrvs didn't exist in the AddrNew, then disable the haproxySrv entry
			(*haproxySrvs)[i].Address = ""
			(*haproxySrvs)[i].Modified = true
			disabled = append(disabled, srv)
		}
	}

	// Configure new Addresses in available HAProxySrvs
	for key, address := range newAddresses {
		if len(disabled) == 0 {
			break
		}
		disabled[0].Address = address.Address
		disabled[0].Modified = true
		disabled[0].Port = address.Port
		disabled = disabled[1:]
		delete(newAddresses, key)
	}
	// Dynamically updates HAProxy backend servers  with HAProxySrvs content
	var addrErr, stateErr error
	for _, srv := range *haproxySrvs {
		if !srv.Modified {
			continue
		}
		if srv.Address == "" {
			// logger.Tracef("server '%s/%s' changed status to %v", newEndpoints.BackendName, srv.Name, "maint")
			addrErr = c.SetServerAddr(BackendName, srv.Name, "127.0.0.1", 0)
			stateErr = c.SetServerState(BackendName, srv.Name, "maint")
		} else {
			// logger.Tracef("server '%s/%s' changed status to %v", newEndpoints.BackendName, srv.Name, "ready")
			addrErr = c.SetServerAddr(BackendName, srv.Name, srv.Address, int(srv.Port))
			stateErr = c.SetServerState(BackendName, srv.Name, "ready")
		}
		if addrErr != nil || stateErr != nil {
			//newEndpoints.DynUpdateFailed = true
			errors.Add(addrErr)
			errors.Add(stateErr)
		}
	}
	return errors.Result()
}
