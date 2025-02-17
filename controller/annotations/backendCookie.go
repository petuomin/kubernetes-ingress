package annotations

import (
	"fmt"
	"strings"

	"github.com/haproxytech/client-native/v2/models"
)

type BackendCookie struct {
	name       string
	cookieName string
	backend    *models.Backend
}

func NewBackendCookie(n string, b *models.Backend) *BackendCookie {
	return &BackendCookie{name: n, backend: b}
}

func (a *BackendCookie) GetName() string {
	return a.name
}

func (a *BackendCookie) Parse(input string) error {
	if len(strings.Fields(input)) != 1 {
		return fmt.Errorf("cookie-persistence: Incorrect input %s", input)
	}
	a.cookieName = input
	return nil
}

func (a *BackendCookie) Update() error {
	if a.cookieName == "" {
		a.backend.Cookie = nil
		return nil
	}
	cookie := models.Cookie{
		Name:     &a.cookieName,
		Type:     "insert",
		Nocache:  true,
		Indirect: true,
	}
	a.backend.Cookie = &cookie
	return nil
}
