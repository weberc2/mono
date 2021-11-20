package main

import (
	"errors"
	"fmt"
	html "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	pz "github.com/weberc2/httpeasy"
)

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
	// create a struct for templating and logging
	x := struct {
		// This is intended to be a user-facing message shown in the templated
		// UI.
		ErrorMessage string `json:"errorMessage"`
		FormAction   string `json:"formAction"`
	}{
		FormAction: ws.BaseURL + "login?" + url.Values{
			"location": []string{r.URL.Query().Get("location")},
		}.Encode(),
	}

	return pz.Ok(pz.HTMLTemplate(ws.LoginForm, &x), &x)
}

func (ws *WebServer) LoginHandler(r pz.Request) pz.Response {
	username, password, err := parseMultiPartForm(r)
	if err != nil {
		return pz.BadRequest(
			pz.Stringf("parsing credentials: %v", err),
			struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}{
				Message: "parsing credentials from multi-part form",
				Error:   err.Error(),
			},
		)
	}
	tokenDetails, err := ws.AuthService.Login(&Credentials{
		User:     UserID(username),
		Password: password,
	})
	if err != nil {
		if errors.Is(err, ErrCredentials) {
			return pz.Unauthorized(
				pz.HTMLTemplate(ws.LoginForm, struct {
					Location     html.HTML
					FormAction   string
					ErrorMessage string
				}{
					FormAction:   ws.BaseURL + "login",
					ErrorMessage: "Invalid credentials",
				}),
				struct {
					User    string `json:"user"`
					Message string `json:"message"`
				}{
					User:    username,
					Message: "login failed",
				},
			)
		}
		return pz.InternalServerError(struct {
			User    string `json:"user"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}{
			User:    username,
			Message: "logging in",
			Error:   err.Error(),
		})
	}

	type parameter struct {
		Value           string   `json:"value"`
		ParseError      string   `json:"parseError,omitempty"`
		ValidationError string   `json:"validationError,omitempty"`
		RedirectDomain  string   `json:"redirectDomain"`
		Parsed          *url.URL `json:"parsed,omitempty"`
	}

	location := r.URL.Query().Get("location")
	logging := struct {
		Message                      string    `json:"message,omitempty"`
		LocationQueryStringParameter parameter `json:"locationQueryStringParameter"`
		Location                     string    `json:"location,omitempty"`
	}{
		LocationQueryStringParameter: parameter{
			Value:          location,
			RedirectDomain: ws.RedirectDomain,
		},
		Location: location,
	}

	if location != "" {
		u, err := url.Parse(location)
		if err != nil {
			logging.LocationQueryStringParameter.ParseError = err.Error()
			logging.Message = "`location` query string parameter is not a " +
				"valid URL"
			return pz.BadRequest(
				pz.String("Invalid redirect value"),
				&logging,
			)
		}
		logging.LocationQueryStringParameter.Parsed = u

		// Make sure the `Host` is either an exact match for the
		// `RedirectDomain` or a valid subdomain. If it's not, then redirect to
		// the default redirect domain.
		if u.Host != ws.RedirectDomain && !strings.HasSuffix(
			u.Host,
			// Note that we have to prepend a `.` onto the `RedirectDomain`
			// before checking if it is a suffix match to be sure we're only
			// matching subdomains. For example, if `RedirectDomain` is
			// `google.com`, an attacker could register `evilgoogle.com` which
			// would match if we didn't prepend the `.` (causing us to send the
			// attacker our tokens).
			fmt.Sprintf(".%s", ws.RedirectDomain),
		) {
			logging.LocationQueryStringParameter.ValidationError = "" +
				"`location` query string parameter host is neither the " +
				"redirect domain itself nor a subdomain thereof. Falling " +
				"back to default URL."
			logging.Location = ws.DefaultRedirectLocation
			location = ws.DefaultRedirectLocation
		}
	} else {
		logging.LocationQueryStringParameter.ValidationError = "`location` " +
			"query string parameter is empty or unset. Falling back to " +
			"default URL."
		logging.Location = ws.DefaultRedirectLocation
		location = ws.DefaultRedirectLocation
	}

	// Previously we used 307 Temporary Redirect, but since we're handling a
	// POST request, the redirect also issued a POST request instead of a GET
	// request. It seems like 303 See Other does what we want.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Redirections#temporary_redirections
	return pz.SeeOther(location, &logging).WithCookies(&http.Cookie{
		Name:     "Access-Token",
		Value:    tokenDetails.AccessToken,
		Domain:   ws.RedirectDomain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}, &http.Cookie{
		Name:     "Refresh-Token",
		Value:    tokenDetails.RefreshToken,
		Domain:   ws.RedirectDomain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
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
