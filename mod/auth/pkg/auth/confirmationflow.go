package auth

import (
	"errors"
	"fmt"
	html "html/template"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/mod/auth/pkg/auth/types"

	. "github.com/weberc2/mono/mod/auth/pkg/prelude"
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
	ackPage, err := ackPage(params.activity)
	if err != nil {
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
		field{ID: "password", Type: "password", Label: "Password"},
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
				return pz.Accepted(pz.String(ackPage))
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
		Handler: formHandler(form.template, func(f url.Values) pz.Response {
			user, err := form.callback(&ws.AuthService, f)
			if err != nil {
				// In certain cases, the callback might want to succeed with
				// logging information, for example, a form for resetting a
				// user's password might want to report success to the user
				// while logging that the provided username didn't exist. To
				// do so, they can return a `*successSentinelErr`.
				var errWithLogging *successSentinelErr
				if errors.As(err, &errWithLogging) {
					return form.success(ws).WithLogging(&logging{
						Message:   errWithLogging.message,
						User:      user,
						ErrorType: reflect.TypeOf(errWithLogging.err).String(),
						Error:     errWithLogging.err.Error(),
					})
				}

				httpErr := &pz.HTTPError{
					Status:  http.StatusInternalServerError,
					Message: "internal server error",
				}

				// return value doesn't matter
				_ = errors.As(err, &httpErr)

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

type successSentinelErr struct {
	message string
	err     error
}

func (err *successSentinelErr) Error() string {
	panic("`(*successSentinelErr).Error() should never be invoked!")
}

type callback func(*AuthService, url.Values) (types.UserID, error)

func ackPage(activity string) (string, error) {
	var sb strings.Builder
	if err := ackPageTemplate.Execute(
		&sb,
		&struct{ Activity string }{activity},
	); err != nil {
		return "", fmt.Errorf("executing ack page template: %w", err)
	}
	return sb.String(), nil
}

var ackPageTemplate = Must(template(`<html>
<head><title>Initiated {{.Activity}}</title></head>
<body>
<h1 id="title">Initiated {{.Activity}}</h1>
<p>An email has been sent to the email address provided. Please check your
email for a confirmation link.</p>
</body>
</html>`))
