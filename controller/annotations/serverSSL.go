package annotations

import (
	"github.com/haproxytech/client-native/v2/models"

	"github.com/haproxytech/kubernetes-ingress/controller/utils"
)

type ServerSSL struct {
	name    string
	enabled bool
	server  *models.Server
}

func NewServerSSL(n string, s *models.Server) *ServerSSL {
	return &ServerSSL{name: n, server: s}
}

func (a *ServerSSL) GetName() string {
	return a.name
}

func (a *ServerSSL) Parse(input string) error {
	var err error
	a.enabled, err = utils.GetBoolValue(input, "server-ssl")
	return err
}

func (a *ServerSSL) Update() error {
	if a.enabled {
		a.server.Ssl = "enabled"
		a.server.Alpn = "h2,http/1.1"
		a.server.Verify = "none"
	} else {
		a.server.Ssl = ""
		a.server.Alpn = ""
		a.server.Verify = ""
	}
	return nil
}
