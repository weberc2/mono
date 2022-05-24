package auth

import (
	"errors"
	"fmt"
	html "html/template"
	"net/http"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/pkg/auth/types"
)

const (
	pathPasswordReset             = "/password/reset"
	pathPasswordResetConfirmation = "/password/confirm-reset"
	pageInitiatedPasswordReset    = `<html>
<head>
	<title>Initiated Password Reset</title>
<body>
<h1>Initiated Password Reset</h1>
<p>An email has been sent to the email address corresponding to the provided
username. Please check your email for a confirmation link.</p>
</body>
</head>
</html>`
)

var (
	routePasswordResetForm = formRoute(
		pathPasswordReset,
		templatePasswordResetForm,
		func(r pz.Request, path string) interface{} {
			return struct {
				FormAction string `json:"formAction"`

				// Need to include this for the template even though we don't
				// pass anything (if this isn't present, the template will
				// error at runtime).
				ErrorMessage string `json:"errorMessage,omitempty"`
			}{FormAction: path}
		},
	)

	routePasswordResetConfirmationForm = formRoute(
		pathPasswordResetConfirmation,
		templatePasswordResetConfirmationForm,
		func(r pz.Request, path string) interface{} {
			return struct {
				FormAction string `json:"formAction"`
				Token      string `json:"-"` // don't log the token (security)

				// Need to include this for the template even though we don't
				// pass anything (if this isn't present, the template will
				// error at runtime).
				ErrorMessage string `json:"errorMessage,omitempty"`
			}{
				FormAction: path,
				Token:      r.URL.Query().Get("t"),
			}
		},
	)

	templatePasswordResetForm = mustHTML(`<html>
<head>
	<title>Password Reset</title>
</head>
<body>
<h1>Password Reset</h1>
{{ if .ErrorMessage }}<p id="error-message">{{ .ErrorMessage }}</p>{{ end }}
<form action="{{ .FormAction }}" method="POST">
	<label for="username">Username</label>
	<input type="text" id="username" name="username"><br><br>
	<input type="submit" value="Submit">
</form>
</body>
</html>`)

	templatePasswordResetConfirmationForm = mustHTML(`<html>
<head>
	<title>Confirm Password Reset</title>
</head>
<body>
<h1>Confirm Password Reset</h1>
{{ if .ErrorMessage }}<p id="error-message">{{ .ErrorMessage }}</p>{{ end }}
<form action="{{ .FormAction }}" method="POST">
	<label for="password">Password</label>
	<input type="password" id="password" name="password"><br><br>
	<input type="hidden" id="token" name="token" value="{{.Token}}">
	<input type="submit" value="Submit">
</form>
</body>
</html>`)
)

func formRoute(
	path string,
	template *html.Template,
	context func(r pz.Request, path string) interface{},
) pz.Route {
	return pz.Route{
		Method: "GET",
		Path:   path,
		Handler: func(r pz.Request) pz.Response {
			ctx := context(r, path)
			return pz.Ok(pz.HTMLTemplate(template, ctx), ctx)
		},
	}
}

func routePasswordResetHandler(ws *WebServer) pz.Route {
	return pz.Route{
		Method: "POST",
		Path:   pathPasswordReset,
		Handler: func(r pz.Request) pz.Response {
			form, err := parseForm(r)
			if err != nil {
				return pz.HandleError(
					"error parsing form data",
					ErrParsingFormData,
					&logging{
						Message:   "parsing password reset form data",
						ErrorType: fmt.Sprintf("%T", err),
						Error:     err.Error(),
					},
				)
			}

			username := types.UserID(form.Get("username"))
			if err := ws.AuthService.ForgotPassword(username); err != nil {
				httpErr := &pz.HTTPError{
					Status:  http.StatusInternalServerError,
					Message: "internal server error",
				}
				errors.As(err, &httpErr)

				// Don't serve error information if the error is 404--we don't
				// want to leak information about whether or not a username
				// exists to potential attackers.
				if httpErr.Status == http.StatusNotFound {
					return pz.Accepted(
						pz.String(pageInitiatedPasswordReset),
						&logging{
							Message: "username not found, but reporting 202 " +
								"Accepted to caller",
							User:      username,
							ErrorType: fmt.Sprintf("%T", err),
							Error:     err.Error(),
						},
					)
				}
				ctx := struct {
					FormAction string `json:"formAction"`

					// for html template
					ErrorMessage string `json:"errorMessage,omitempty"`

					// logging only
					PrivateError string `json:"privateError,omitempty"`
				}{
					FormAction:   ws.BaseURL + pathPasswordReset,
					ErrorMessage: httpErr.Message,
					PrivateError: err.Error(),
				}
				return pz.Response{
					Status: httpErr.Status,
					Data:   pz.HTMLTemplate(templatePasswordResetForm, &ctx),
				}.WithLogging(&ctx)
			}
			return pz.Accepted(
				pz.String(pageInitiatedPasswordReset),
				&logging{
					Message: "kicked off password reset",
					User:    username,
				},
			)
		},
	}
}

func routePasswordResetConfirmationHandler(ws *WebServer) pz.Route {
	return pz.Route{
		Method: "POST",
		Path:   pathPasswordResetConfirmation,
		Handler: func(r pz.Request) pz.Response {
			form, err := parseForm(r)
			if err != nil {
				return pz.HandleError(
					"error parsing form data",
					ErrParsingFormData,
					&logging{
						Message:   "parsing password reset confirmation form",
						ErrorType: fmt.Sprintf("%T", err),
						Error:     err.Error(),
					},
				)
			}

			user, err := ws.AuthService.UpdatePassword(&UpdatePassword{
				Token:    form.Get("token"),
				Password: form.Get("password"),
			})
			if err != nil {
				httpErr := &pz.HTTPError{
					Status:  http.StatusInternalServerError,
					Message: "internal server error",
				}
				_ = errors.As(err, &httpErr)
				ctx := struct {
					FormAction string `json:"formAction"`

					// The user whose password we're attempting to update. May
					// be empty if the token is invalid.
					User types.UserID `json:"user"`

					// hidden form field
					Token string `json:"-"` // don't log secrets

					// for html template
					ErrorMessage string `json:"errorMessage,omitempty"`

					// logging only
					PrivateError string `json:"privateError,omitempty"`

					// Type of PrivateError
					ErrorType string `json:"errorType,omitempty"`
				}{
					FormAction:   pathRegistrationConfirmationForm,
					User:         user,
					Token:        form.Get("token"),
					ErrorMessage: httpErr.Message,
					PrivateError: err.Error(),
					ErrorType:    fmt.Sprintf("%T", err),
				}
				return pz.Response{
					Status: httpErr.Status,
					Data: pz.HTMLTemplate(
						templatePasswordResetConfirmationForm,
						&ctx,
					),
				}.WithLogging(&ctx)
			}
			return pz.SeeOther(ws.DefaultRedirectLocation, &struct {
				User    types.UserID `json:"user"`
				Message string       `json:"message"`
			}{
				User:    user,
				Message: "successfully reset user password",
			})
		},
	}
}
