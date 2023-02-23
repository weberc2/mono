package comments

import (
	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/mod/auth/pkg/auth/client"
)

type Auth struct {
	client.AuthType
	client.Authenticator
}

func (a *Auth) Auth(r pz.Route) pz.Route {
	return pz.Route{
		Method:  r.Method,
		Path:    r.Path,
		Handler: a.Authenticator.Auth(a.AuthType, r.Handler),
	}
}

func (a *Auth) Optional(r pz.Route) pz.Route {
	return pz.Route{
		Method:  r.Method,
		Path:    r.Path,
		Handler: a.Authenticator.Optional(a.AuthType, r.Handler),
	}
}
