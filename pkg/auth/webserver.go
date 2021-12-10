package auth

import (
	"errors"
	"fmt"
	html "html/template"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/weberc2/auth/pkg/types"
	pz "github.com/weberc2/httpeasy"
)

type Code struct {
	Code string `json:"code"`
}

// WebServer serves the authentication pages for websites (as opposed to
// single-page apps). It passes tokens to the client via cookies.
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

	// LoginForm is the HTML template containing the login form page HTML.
	LoginForm *html.Template
}

func (ws *WebServer) LoginFormPage(r pz.Request) pz.Response {
	query := r.URL.Query()

	// create a struct for templating and logging
	x := struct {
		// This is intended to be a user-facing message shown in the templated
		// UI.
		ErrorMessage string `json:"errorMessage"`
		FormAction   string `json:"formAction"`
	}{
		FormAction: ws.BaseURL + "login?" + url.Values{
			"callback": []string{query.Get("callback")},
			"redirect": []string{query.Get("redirect")},
		}.Encode(),
	}

	return pz.Ok(pz.HTMLTemplate(ws.LoginForm, &x), &x)
}

func (ws *WebServer) LoginHandler(r pz.Request) pz.Response {
	username, password, err := parseMultiPartForm(r)
	if err != nil {
		return pz.BadRequest(
			pz.Stringf("parsing credentials: %v", err),
			&logging{
				Message: "parsing credentials from multi-part form",
				Error:   err.Error(),
			},
		)
	}

	code, err := ws.AuthService.LoginAuthCode(&types.Credentials{
		User:     types.UserID(username),
		Password: password,
	})
	if err != nil {
		if errors.Is(err, ErrCredentials) {
			return pz.Unauthorized(
				pz.HTMLTemplate(ws.LoginForm, &struct {
					Location     html.HTML
					FormAction   string
					ErrorMessage string
				}{
					FormAction:   ws.BaseURL + "login",
					ErrorMessage: "Invalid credentials",
				}),
				&logging{
					User:    types.UserID(username),
					Message: "login failed",
					Error:   err.Error(),
				},
			)
		}
		return pz.InternalServerError(&logging{
			User:      types.UserID(username),
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

func parseMultiPartForm(r pz.Request) (string, string, error) {
	// Read at most 2kb data to avoid DOS attack. That should be plenty for our
	// form.
	data, err := ioutil.ReadAll(io.LimitReader(r.Body, 2056))
	if err != nil {
		return "", "", fmt.Errorf("reading request body: %w", err)
	}
	form, err := url.ParseQuery(string(data))
	if err != nil {
		return "", "", fmt.Errorf("parsing form data: %w", err)
	}
	return form.Get("username"), form.Get("password"), nil
}
