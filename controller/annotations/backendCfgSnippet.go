package annotations

import (
	"errors"
	"strings"

	"github.com/haproxytech/client-native/v2/models"

	"github.com/haproxytech/kubernetes-ingress/controller/haproxy/api"
)

type BackendCfgSnippet struct {
	name    string
	data    []string
	client  api.HAProxyClient
	backend *models.Backend
}

func NewBackendCfgSnippet(n string, c api.HAProxyClient, b *models.Backend) *BackendCfgSnippet {
	return &BackendCfgSnippet{name: n, client: c, backend: b}
}

func (a *BackendCfgSnippet) GetName() string {
	return a.name
}

func (a *BackendCfgSnippet) Parse(input string) error {
	for _, line := range strings.Split(strings.Trim(input, "\n"), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			a.data = append(a.data, line)
		}
	}
	if len(a.data) == 0 {
		return errors.New("unable to parse config-snippet: empty input")
	}
	return nil
}

func (a *BackendCfgSnippet) Update() error {
	if len(a.data) == 0 {
		return a.client.BackendCfgSnippetSet(a.backend.Name, nil)
	}
	return a.client.BackendCfgSnippetSet(a.backend.Name, &a.data)
}
