package annotations

import (
	"errors"
	"strings"

	"github.com/haproxytech/config-parser/v4/types"
	"github.com/haproxytech/kubernetes-ingress/controller/haproxy/api"
)

type GlobalCfgSnippet struct {
	name string
	// data   *types.StringSliceC
	data   []string
	client api.HAProxyClient
}

func NewGlobalCfgSnippet(n string, c api.HAProxyClient) *GlobalCfgSnippet {
	return &GlobalCfgSnippet{name: n, client: c}
}

func (a *GlobalCfgSnippet) GetName() string {
	return a.name
}

func (a *GlobalCfgSnippet) Parse(input string) error {
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

func (a *GlobalCfgSnippet) Update() error {
	if len(a.data) == 0 {
		logger.Infof("Removing global config-snippet")
		return a.client.GlobalCfgSnippet(nil)
	}
	logger.Infof("Updating global config-snippet")
	return a.client.GlobalCfgSnippet(&types.StringSliceC{Value: a.data})
}
