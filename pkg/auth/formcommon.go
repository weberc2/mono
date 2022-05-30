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

const (
	pathRegistration              = "/register"
	pathRegistrationConfirmation  = "/confirm"
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
	pageInitiatedRegistration = `<html>
<head>
	<title>Registration Accepted</title>
<body>
<h1>
Registration Accepted
</h1>
<p>An email has been sent to the email address provided. Please check your
email for a confirmation link.</p>
</body>
</head>
</html>`
)

var (
	routePasswordResetForm = formRoute(
		pathPasswordReset,
		templatePasswordResetForm,
	)

	routePasswordResetConfirmationForm = formRoute(
		pathPasswordResetConfirmation,
		templatePasswordResetConfirmationForm,
	)

	routeRegistrationForm = formRoute(
		pathRegistration,
		templateRegistrationForm,
	)

	routeRegistrationConfirmationForm = formRoute(
		pathRegistrationConfirmation,
		templateRegistrationConfirmationForm,
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

	templateRegistrationForm = mustHTML(
		`<html>
<head>
	<title>Register</title>
</head>
<body>
<h1>Register</h1>
{{ if .ErrorMessage }}<p id="error-message">{{ .ErrorMessage }}</p>{{ end }}
<form action="{{ .FormAction }}" method="POST">
	<label for="username">Username</label>
	<input type="text" id="username" name="username"><br><br>
	<label for="email">Email</label>
	<input type="text" id="email" name="email"><br><br>
	<input type="submit" value="Submit">
</form>
</body>
</html>`)

	templateRegistrationConfirmationForm = mustHTML(`<html>
<head>
	<title>Confirm Registration</title>
</head>
<body>
<h1>Confirm Registration<h1>
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
) pz.Route {
	return pz.Route{
		Method: "GET",
		Path:   path,
		Handler: func(r pz.Request) pz.Response {
			ctx := struct {
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
			return pz.Ok(pz.HTMLTemplate(template, ctx), ctx)
		},
	}
}

func routeRegistrationHandler(ws *WebServer) pz.Route {
	return routeHandler(ws, &handlerParams{
		activity:    "registration",
		path:        pathRegistration,
		successPage: pageInitiatedRegistration,
		template:    templateRegistrationForm,
		callback: func(auth *AuthService, f url.Values) (types.UserID, error) {
			return auth.Register(
				types.UserID(f.Get("username")),
				f.Get("email"),
			)
		},
	})
}

func routePasswordResetHandler(ws *WebServer) pz.Route {
	return routeHandler(ws, &handlerParams{
		activity:    "password reset",
		path:        pathPasswordReset,
		successPage: pageInitiatedPasswordReset,
		template:    templatePasswordResetForm,
		callback: func(auth *AuthService, f url.Values) (types.UserID, error) {
			return auth.ForgotPassword(types.UserID(f.Get("username")))
		},
	})
}

type handlerParams struct {
	activity    string
	path        string
	successPage string
	template    *html.Template
	callback    func(*AuthService, url.Values) (types.UserID, error)
}

func routeHandler(ws *WebServer, params *handlerParams) pz.Route {
	return pz.Route{
		Method: "POST",
		Path:   params.path,
		Handler: func(r pz.Request) pz.Response {
			form, err := parseForm(r)
			if err != nil {
				return pz.HandleError(
					"error parsing form data",
					ErrParsingFormData,
					&logging{
						Message: fmt.Sprintf(
							"%s: parsing form data",
							params.activity,
						),
						ErrorType: fmt.Sprintf("%T", err),
						Error:     err.Error(),
					},
				)
			}

			user, err := params.callback(&ws.AuthService, form)
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
					return pz.Accepted(
						pz.String(params.successPage),
						&logging{
							Message: "username not found, but reporting 202 " +
								"Accepted to caller",
							User:      user,
							ErrorType: fmt.Sprintf("%T", err),
							Error:     err.Error(),
						},
					)
				}
				ctx := formContext{
					FormAction:   ws.BaseURL + params.path,
					ErrorMessage: httpErr.Message,
					PrivateError: err.Error(),
				}
				return pz.Response{
					Status: httpErr.Status,
					Data:   pz.HTMLTemplate(params.template, &ctx),
				}.WithLogging(&ctx)
			}
			return pz.Accepted(
				pz.String(params.successPage),
				&logging{
					Message: fmt.Sprintf("kicked off %s", params.activity),
					User:    user,
				},
			)
		},
	}
}

type formContext struct {
	FormAction string `json:"formAction"`

	// for html template
	ErrorMessage string `json:"errorMessage,omitempty"`

	// logging only
	PrivateError string `json:"privateError,omitempty"`
}

func routeRegistrationConfirmationHandler(ws *WebServer) pz.Route {
	return routeConfirmation(ws, &confirmationParams{
		activity: "registration",
		path:     pathRegistrationConfirmation,
		create:   true,
	})
}

func routePasswordResetConfirmationHandler(ws *WebServer) pz.Route {
	return routeConfirmation(ws, &confirmationParams{
		activity: "password reset",
		path:     pathPasswordResetConfirmation,
		create:   false,
	})
}

func routeConfirmation(ws *WebServer, params *confirmationParams) pz.Route {
	return pz.Route{
		Method: "POST",
		Path:   params.path,
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
				Create:   params.create,
				Token:    form.Get("token"),
				Password: form.Get("password"),
			})
			if err != nil {
				var httpErr pz.Error = &pz.HTTPError{
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
					FormAction:   params.path,
					User:         user,
					Token:        form.Get("token"),
					ErrorMessage: httpErr.HTTPError().Message,
					PrivateError: err.Error(),
					ErrorType:    fmt.Sprintf("%T", err),
				}
				return pz.Response{
					Status: httpErr.HTTPError().Status,
					Data:   pz.HTMLTemplate(params.template, &ctx),
				}.WithLogging(&ctx)
			}
			return pz.SeeOther(ws.DefaultRedirectLocation, &struct {
				User    types.UserID `json:"user"`
				Message string       `json:"message"`
			}{
				User:    user,
				Message: fmt.Sprintf("%s: success", params.activity),
			})
		},
	}
}

var ErrParsingFormData = &pz.HTTPError{
	Status:  http.StatusBadRequest,
	Message: "error parsing form data",
}

type confirmationParams struct {
	activity string
	path     string
	create   bool
	template *html.Template
}
