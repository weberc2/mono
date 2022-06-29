package auth

import (
	"errors"
	"fmt"
	html "html/template"
	"net/url"
	"strings"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/pkg/auth/types"

	. "github.com/weberc2/mono/pkg/prelude"
)

func loginHandler(ws *WebServer) pz.Handler {
	return func(r pz.Request) pz.Response {
		return handleForm(loginForm, r, func(form url.Values) pz.Response {
			return handleLogin(ws, r, form)
		})
	}
}

func handleLogin(ws *WebServer, r pz.Request, form url.Values) pz.Response {
	username := types.UserID(form.Get("username"))

	code, err := ws.AuthService.LoginAuthCode(&types.Credentials{
		User:     username,
		Password: form.Get("password"),
	})
	if err != nil {
		if errors.Is(err, ErrCredentials) {
			return pz.Unauthorized(
				pz.HTMLTemplate(loginForm, &struct {
					Location     html.HTML
					FormAction   string
					ErrorMessage string
				}{
					// Not only do we need the correct path, but we also
					// need to preserve the query params (e.g.,
					// `callback`).
					FormAction:   r.URL.String(),
					ErrorMessage: "Invalid credentials",
				}),
				&logging{
					User:    username,
					Message: "login failed",
					Error:   err.Error(),
				},
			)
		}
		return pz.InternalServerError(&logging{
			User:      username,
			Message:   "logging in",
			ErrorType: fmt.Sprintf("%T", err),
			Error:     err.Error(),
		})
	}

	query := r.URL.Query()
	context := struct {
		Message  string `json:"message,omitempty"`
		Target   string `json:"target,omitempty"`
		Redirect redirectResult
		Callback redirectResult
	}{
		Callback: redirectResult{
			Specified: query.Get("callback"),
			Default:   ws.DefaultRedirectLocation,
			Domain:    ws.RedirectDomain,
		},
		Redirect: redirectResult{
			Specified: query.Get("redirect"),
			Default:   ws.DefaultRedirectLocation,
			Domain:    ws.RedirectDomain,
		},
	}

	validateRedirect(&context.Callback)
	if context.Callback.ParseError != "" {
		context.Message = "`callback` parameter contains invalid URL"
		return pz.BadRequest(nil, &context)
	}
	validateRedirect(&context.Redirect)
	if context.Redirect.ParseError != "" {
		context.Message = "`redirect` parameter contains invalid URL"
		return pz.BadRequest(nil, &context)
	}

	context.Target = context.Callback.Actual + "?" + url.Values{
		"code":     []string{code},
		"redirect": []string{context.Redirect.Actual},
		"callback": []string{context.Callback.Actual},
	}.Encode()

	// Previously we used 307 Temporary Redirect, but since we're handling a
	// POST request, the redirect also issued a POST request instead of a GET
	// request. It seems like 303 See Other does what we want.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Redirections#temporary_redirections
	return pz.SeeOther(context.Target, &context)
}

func validateRedirect(context *redirectResult) {
	if context.Specified != "" {
		u, err := url.Parse(context.Specified)
		if err != nil {
			context.ParseError = err.Error()
			return
		}
		context.Parsed = u

		// Make sure the `Host` is either an exact match for the
		// `RedirectDomain` or a valid subdomain. If it's not, then redirect to
		// the default redirect domain.
		if u.Host != context.Domain && !strings.HasSuffix(
			u.Host,
			// Note that we have to prepend a `.` onto the `RedirectDomain`
			// before checking if it is a suffix match to be sure we're only
			// matching subdomains. For example, if `RedirectDomain` is
			// `google.com`, an attacker could register `evilgoogle.com` which
			// would match if we didn't prepend the `.` (causing us to send the
			// attacker our tokens).
			fmt.Sprintf(".%s", context.Domain),
		) {
			context.ValidationError = "" +
				"`location` query string parameter host is neither the " +
				"redirect domain itself nor a subdomain thereof. Falling " +
				"back to default URL."
			context.Actual = context.Default
		}
	} else {
		context.ValidationError = "`location` query string " +
			"parameter is empty or unset. Falling back to " +
			"default URL."
		context.Actual = context.Default
	}
	context.Actual = context.Specified
}

type redirectResult struct {
	Domain          string   `json:"domain"`
	Specified       string   `json:"specified"`
	Default         string   `json:"default"`
	Parsed          *url.URL `json:"parsed,omitempty"`
	Actual          string   `json:"actual,omitempty"`
	ParseError      string   `json:"parseError,omitempty"`
	ValidationError string   `json:"validationError,omitempty"`
}

func loginFormHandler(baseURL string) pz.Handler {
	return func(r pz.Request) pz.Response {
		query := r.URL.Query()

		// create a struct for templating and logging
		context := struct {
			FormAction   string `json:"formAction"`
			ErrorMessage string `json:"-"`
		}{
			FormAction: fmt.Sprintf(
				"%slogin?%s",
				baseURL,
				url.Values{
					"callback": []string{query.Get("callback")},
					"redirect": []string{query.Get("redirect")},
				}.Encode(),
			),
		}

		return pz.Ok(pz.HTMLTemplate(loginForm, &context), &context)
	}
}

var (
	loginForm = Must(formHTMLNoEscape(
		"Login",
		"{{.FormAction}}",
		field{ID: "username", Label: "Username"},
		field{ID: "password", Type: "password", Label: "Password"},
	))
)
