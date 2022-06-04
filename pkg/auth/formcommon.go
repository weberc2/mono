package auth

import (
	"errors"
	"fmt"
	html "html/template"
	"net/http"
	"net/url"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/pkg/auth/types"
)

type confirmationFlow struct {
	activity     string
	main         form
	confirmation form
}

func (flow *confirmationFlow) routeMainForm() pz.Route {
	return flow.main.formRoute()
}

func (flow *confirmationFlow) routeConfirmationForm() pz.Route {
	return flow.confirmation.formRoute()
}

func (flow *confirmationFlow) routeMainHandler(ws *WebServer) pz.Route {
	return flow.main.handlerRoute(flow.activity, ws)
}

func (flow *confirmationFlow) routeConfirmationHandler(
	ws *WebServer,
) pz.Route {
	return flow.confirmation.handlerRoute(flow.activity, ws)
}

type form struct {
	path     string
	template *html.Template
	callback callback
	success  success
}

func (form *form) formRoute() pz.Route {
	return pz.Route{
		Method: "GET",
		Path:   form.path,
		Handler: func(r pz.Request) pz.Response {
			ctx := struct {
				FormAction string `json:"formAction"`
				Token      string `json:"-"` // don't log the token (security)

				// Need to include this for the template even though we don't
				// pass anything (if this isn't present, the template will
				// error at runtime).
				ErrorMessage string `json:"errorMessage,omitempty"`
			}{
				FormAction: form.path,
				Token:      r.URL.Query().Get("t"),
			}
			return pz.Ok(pz.HTMLTemplate(form.template, ctx), ctx)
		},
	}
}

func (form *form) handlerRoute(activity string, ws *WebServer) pz.Route {
	return pz.Route{
		Method: "POST",
		Path:   form.path,
		Handler: func(r pz.Request) pz.Response {
			f, err := parseForm(r)
			if err != nil {
				return pz.HandleError(
					"error parsing form data",
					ErrParsingFormData,
					&logging{
						Message: fmt.Sprintf(
							"%s: parsing form data",
							activity,
						),
						ErrorType: fmt.Sprintf("%T", err),
						Error:     err.Error(),
					},
				)
			}

			user, err := form.callback(&ws.AuthService, f)
			if err != nil {
				httpErr := &pz.HTTPError{
					Status:  http.StatusInternalServerError,
					Message: "internal server error",
				}

				// return value doesn't matter
				_ = errors.As(err, &httpErr)

				// Don't serve error information if the error is 404--we don't
				// want to leak information about whether or not a username
				// exists to potential attackers. A not-found error should only
				// be returned when updating a user password (user-not-found
				// is the happy path for user registration).
				if httpErr.Status == http.StatusNotFound {
					return form.success(ws).WithLogging(&logging{
						Message: "username not found, but reporting 202 " +
							"Accepted to caller",
						User:      user,
						ErrorType: fmt.Sprintf("%T", err),
						Error:     err.Error(),
					})
				}
				ctx := struct {
					FormAction string `json:"formAction"`

					// for html template
					ErrorMessage string `json:"errorMessage,omitempty"`

					// logging only
					PrivateError string `json:"privateError,omitempty"`
				}{
					FormAction:   ws.BaseURL + form.path,
					ErrorMessage: httpErr.Message,
					PrivateError: err.Error(),
				}
				return pz.Response{
					Status: httpErr.Status,
					Data:   pz.HTMLTemplate(form.template, &ctx),
				}.WithLogging(&ctx)
			}
			return form.success(ws).WithLogging(&logging{
				Message: fmt.Sprintf("%s: success", activity),
				User:    user,
			})
		},
	}
}

var ErrParsingFormData = &pz.HTTPError{
	Status:  http.StatusBadRequest,
	Message: "error parsing form data",
}

func successDefaultRedirect(ws *WebServer) pz.Response {
	return pz.SeeOther(ws.DefaultRedirectLocation)
}

func successAccepted(body string) func(*WebServer) pz.Response {
	return func(*WebServer) pz.Response {
		return pz.Accepted(pz.String(body))
	}
}

type callback func(*AuthService, url.Values) (types.UserID, error)

func callbackUpdatePassword(create bool) callback {
	return func(auth *AuthService, f url.Values) (types.UserID, error) {
		return auth.UpdatePassword(&UpdatePassword{
			Create:   create,
			Token:    f.Get("token"),
			Password: f.Get("password"),
		})
	}
}

type success func(*WebServer) pz.Response
