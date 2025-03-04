package annotations

import (
	"github.com/haproxytech/client-native/v2/models"

	"github.com/haproxytech/kubernetes-ingress/controller/utils"
)

type ServerCheckInter struct {
	name   string
	inter  int64
	server *models.Server
}

func NewServerCheckInter(n string, s *models.Server) *ServerCheckInter {
	return &ServerCheckInter{name: n, server: s}
}

func (a *ServerCheckInter) GetName() string {
	return a.name
}

func (a *ServerCheckInter) Parse(input string) error {
	value, err := utils.ParseTime(input)
	if err != nil {
		return err
	}
	a.inter = *value
	return nil
}

func (a *ServerCheckInter) Update() error {
	a.server.Inter = &a.inter
	return nil
}
