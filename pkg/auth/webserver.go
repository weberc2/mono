package auth

import (
	"errors"
	"net/url"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/pkg/auth/types"
)

type Code struct {
	Code string `json:"code"`
}

// WebServer serves the authentication pages for websites (as opposed to
// single-page apps).
type WebServer struct {
	// AuthService is the authentication service backend.
	AuthService AuthService

	// BaseURL is the base URL for the UIService. This **must** end with a
	// trailing slash.
	BaseURL string

	// RedirectDomain specifies the set of subdomains that we will redirect to.
	// So for example, if the `RedirectDomain` field is `google.com`, then we
	// will redirect to `foo.google.com` but not `attacker.com`.
	RedirectDomain string

	// DefaultRedirectLocation is the destination we send users to on success
	// if no redirect location was specified in the query string.
	DefaultRedirectLocation string
}

func (ws *WebServer) Routes() []pz.Route {
	return []pz.Route{
		ws.LoginFormRoute(),
		ws.LoginHandlerRoute(),
		ws.RegistrationFormRoute(),
		ws.RegistrationHandlerRoute(),
		ws.RegistrationConfirmationFormRoute(),
		ws.RegistrationConfirmationHandlerRoute(),
		ws.PasswordResetFormRoute(),
		ws.PasswordResetHandlerRoute(),
		ws.PasswordResetConfirmationFormRoute(),
		ws.PasswordResetConfirmationHandlerRoute(),
	}
}

func (ws *WebServer) RegistrationFormRoute() pz.Route {
	return flowRegistration.routeMainForm()
}

func (ws *WebServer) RegistrationHandlerRoute() pz.Route {
	return flowRegistration.routeMainHandler(ws)
}

func (ws *WebServer) RegistrationConfirmationFormRoute() pz.Route {
	return flowRegistration.routeConfirmationForm()
}

func (ws *WebServer) RegistrationConfirmationHandlerRoute() pz.Route {
	return flowRegistration.routeConfirmationHandler(ws)
}

func (ws *WebServer) PasswordResetFormRoute() pz.Route {
	return flowPasswordReset.routeMainForm()
}

func (ws *WebServer) PasswordResetHandlerRoute() pz.Route {
	return flowPasswordReset.routeMainHandler(ws)
}

func (ws *WebServer) PasswordResetConfirmationFormRoute() pz.Route {
	return flowPasswordReset.routeConfirmationForm()
}

func (ws *WebServer) PasswordResetConfirmationHandlerRoute() pz.Route {
	return flowPasswordReset.routeConfirmationHandler(ws)
}

func (ws *WebServer) LoginFormRoute() pz.Route {
	return pz.Route{
		Method:  "GET",
		Path:    "/login",
		Handler: loginFormHandler(ws.BaseURL),
	}
}

func (ws *WebServer) LoginHandlerRoute() pz.Route {
	return pz.Route{Path: "/login", Method: "POST", Handler: loginHandler(ws)}
}

var (
	flowRegistration = must(newConfirmationFlow(&confirmationFlowParams{
		activity: "Registration",
		basePath: "/registration",
		fields: []field{
			{ID: "username", Label: "Username"},
			{ID: "email", Label: "Email"},
		},
		mainCallback: func(
			auth *AuthService,
			f url.Values,
		) (types.UserID, error) {
			user := types.UserID(f.Get("username"))
			return user, auth.Register(user, f.Get("email"))
		},
		create: true,
	}))

	flowPasswordReset = must(newConfirmationFlow(&confirmationFlowParams{
		activity: "Password Reset",
		basePath: "/password-reset",
		fields:   []field{{ID: "username", Label: "Username"}},
		mainCallback: func(
			auth *AuthService,
			f url.Values,
		) (types.UserID, error) {
			user := types.UserID(f.Get("username"))
			if err := auth.ForgotPassword(user); err != nil {
				// Don't serve error information if the error is 404--we don't
				// want to leak information about whether or not a username
				// exists to potential attackers. A not-found error should only
				// be returned when updating a user password (user-not-found
				// is the happy path for user registration).
				if errors.Is(err, types.ErrUserNotFound) {
					return user, &successSentinelErr{
						message: "username not found, but reporting 202 " +
							"Accepted to caller",
						err: err,
					}
				}
				return user, err
			}
			return user, nil
		},
		create: false,
	}))
)
