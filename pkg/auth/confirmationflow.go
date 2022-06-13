package auth

import (
	"errors"
	"fmt"
	html "html/template"
	"net/http"
	"net/url"
	"strings"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/pkg/auth/types"
)

type confirmationFlowParams struct {
	activity     string
	basePath     string
	fields       []field
	mainCallback callback
	create       bool
}

func newConfirmationFlow(
	params *confirmationFlowParams,
) (*confirmationFlow, error) {
	var sb strings.Builder
	if err := ackPageTemplate.Execute(
		&sb,
		&struct{ Activity string }{params.activity},
	); err != nil {
		return nil, fmt.Errorf(
			"confirmation flow `%s`: making ack page: %w",
			params.activity,
			err,
		)
	}

	mainTemplate, err := formHTMLEscape(
		params.activity,
		params.basePath,
		params.fields...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"confirmation flow `%s`: making main form template: %w",
			params.activity,
			err,
		)
	}

	confirmationPath := params.basePath + "/confirm"
	confirmationTemplate, err := formHTMLEscape(
		"Confirm "+params.activity,
		confirmationPath,
		field{ID: "password", Label: "Password"},
		field{ID: "token", Hidden: true, Value: "{{.Token}}"},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"confirmation flow `%s`: making confirmation form template: %w",
			params.activity,
			err,
		)
	}

	return &confirmationFlow{
		activity: params.activity,
		main: form{
			path:     params.basePath,
			template: mainTemplate,
			callback: params.mainCallback,
			success: func(*WebServer) pz.Response {
				return pz.Accepted(pz.String(sb.String()))
			},
		},
		confirmation: form{
			path:     confirmationPath,
			template: confirmationTemplate,
			callback: func(
				auth *AuthService,
				f url.Values,
			) (types.UserID, error) {
				return auth.UpdatePassword(&UpdatePassword{
					Create:   params.create,
					Token:    f.Get("token"),
					Password: f.Get("password"),
				})
			},
			success: func(ws *WebServer) pz.Response {
				return pz.SeeOther(ws.DefaultRedirectLocation)
			},
		},
	}, nil
}

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
	success  func(*WebServer) pz.Response
}

func (form *form) formRoute() pz.Route {
	return pz.Route{
		Method: "GET",
		Path:   form.path,
		Handler: func(r pz.Request) pz.Response {
			ctx := formData{Token: r.URL.Query().Get("t")}
			return pz.Ok(pz.HTMLTemplate(form.template, ctx), ctx)
		},
	}
}

func (form *form) handlerRoute(activity string, ws *WebServer) pz.Route {
	return pz.Route{
		Method: "POST",
		Path:   form.path,
		Handler: formHandler(func(f url.Values) pz.Response {
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
				ctx := formData{
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
		}),
	}
}

type callback func(*AuthService, url.Values) (types.UserID, error)

var ackPageTemplate = must(template(`<html>
<head><title>Initiated {{.Activity}}</title></head>
<body>
<h1>Initiated {{.Activity}}</h1>
<p>An email has been sent to the email address provided. Please check your
email for a confirmation link.</p>
</body>
</html>`))

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
