package annotations

import (
	"github.com/haproxytech/client-native/v2/models"

	"github.com/haproxytech/kubernetes-ingress/controller/utils"
)

type BackendTimeoutCheck struct {
	name    string
	timeout *int64
	backend *models.Backend
}

func NewBackendTimeoutCheck(n string, b *models.Backend) *BackendTimeoutCheck {
	return &BackendTimeoutCheck{name: n, backend: b}
}

func (a *BackendTimeoutCheck) GetName() string {
	return a.name
}

func (a *BackendTimeoutCheck) Parse(input string) error {
	var err error
	a.timeout, err = utils.ParseTime(input)
	return err
}

func (a *BackendTimeoutCheck) Update() error {
	a.backend.CheckTimeout = a.timeout
	return nil
}
