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

// UIService serves the authentication pages for websites (as opposed to
// single-page apps). It passes tokens to the client via cookies.
type UIService struct {
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

func (uis *UIService) LoginFormPage(r pz.Request) pz.Response {
	values := r.URL.Query()
	location := values.Get("location")
	if location != "" {
		u, err := url.Parse(location)
		if err != nil {
			return pz.BadRequest(
				pz.String("Invalid redirect value"),
				struct {
					Message, Location, Error string
				}{
					Message: "'location' query string parameter " +
						"is not a valid URL",
					Location: location,
					Error:    err.Error(),
				},
			)
		}

		// Make sure the `Host` is either an exact match for the
		// `RedirectDomain` or a valid subdomain. If it's not, then redirect to
		// the default redirect domain.
		if u.Host != uis.RedirectDomain || !strings.HasSuffix(
			u.Host,
			// Note that we have to prepend a `.` onto the `RedirectDomain`
			// before checking if it is a suffix match to be sure we're only
			// matching subdomains. For example, if `RedirectDomain` is
			// `google.com`, an attacker could register `evilgoogle.com` which
			// would match if we didn't prepend the `.` (causing us to send the
			// attacker our tokens).
			fmt.Sprintf(".%s", uis.RedirectDomain),
		) {
			location = uis.DefaultRedirectLocation
		}
	} else {
		location = uis.DefaultRedirectLocation
	}
	return pz.Ok(pz.HTMLTemplate(uis.LoginForm, struct {
		Location     html.HTML
		FormAction   string
		ErrorMessage string
	}{
		FormAction: uis.BaseURL + "login",
		Location:   html.HTML(location),
	}))
}

func (uis *UIService) LoginHandler(r pz.Request) pz.Response {
	username, password, err := parseMultiPartForm(r)
	if err != nil {
		return pz.BadRequest(
			pz.Stringf("parsing credentials: %v", err),
			struct {
				Message, Error string
			}{
				Message: "parsing credentials from multi-part form",
				Error:   err.Error(),
			},
		)
	}
	tokenDetails, err := uis.AuthService.Login(&Credentials{
		User:     UserID(username),
		Password: password,
	})
	if err != nil {
		if errors.Is(err, ErrCredentials) {
			return pz.Unauthorized(
				pz.HTMLTemplate(uis.LoginForm, struct {
					Location     html.HTML
					FormAction   string
					ErrorMessage string
				}{
					FormAction:   uis.BaseURL + "login",
					ErrorMessage: "Invalid credentials",
				}),
				struct {
					User    UserID
					Message string
				}{
					User:    UserID(username),
					Message: "login failed",
				},
			)
		}
		return pz.InternalServerError(struct {
			Message, Error string
		}{
			Message: "logging in",
			Error:   err.Error(),
		})
	}

	location := r.URL.Query().Get("location")
	if location == "" {
		location = uis.DefaultRedirectLocation
	}
	return pz.TemporaryRedirect(
		location,
		struct {
			Status        string
			RedirectingTo string
		}{
			Status:        "successful",
			RedirectingTo: location,
		},
	).WithCookies(&http.Cookie{
		Name:     "Access-Token",
		Value:    tokenDetails.AccessToken,
		Domain:   uis.RedirectDomain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
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
