package auth

import (
	_ "embed"
	"errors"
	"fmt"
	html "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	pz "github.com/weberc2/httpeasy"

	. "github.com/weberc2/mono/pkg/prelude"
)

type field struct {
	ID     string
	Label  string
	Type   string
	Hidden bool
	Value  string
}

// formHTMLEscape creates a new form with the `action` parameter HTML-escaped.
func formHTMLEscape(
	title string,
	action string,
	fields ...field,
) (*html.Template, error) {
	return formHTMLNoEscape(title, html.HTMLEscapeString(action), fields...)
}

// formHTMLNoEscape creates a new form without HTML-escaping the `action`
// parameter.
func formHTMLNoEscape(
	title string,
	action string,
	fields ...field,
) (*html.Template, error) {
	var sb strings.Builder
	if err := formTemplate.Execute(
		&sb,
		&genericFormData{
			Title:  title,
			Action: html.HTMLAttr("action=" + action),
			Fields: fields,
		},
	); err != nil {
		return nil, fmt.Errorf(
			"form `%s`: generating template: %w",
			title,
			err,
		)
	}

	t, err := html.New("").Parse(sb.String())
	if err != nil {
		return nil, fmt.Errorf(
			"form `%s`: parsing generated template: %w",
			title,
			err,
		)
	}
	return t, nil
}

func formHandler(
	form *html.Template,
	next func(url.Values) pz.Response,
) pz.Handler {
	return func(r pz.Request) pz.Response {
		return handleForm(form, r, next)
	}
}

func handleForm(
	form *html.Template,
	r pz.Request,
	next func(url.Values) pz.Response,
) pz.Response {
	f, err := parseForm(r)
	if err != nil {
		return formError(form, err)
	}
	return next(f)
}

func formError(form *html.Template, err error) pz.Response {
	httpErr := &pz.HTTPError{
		Status:  http.StatusInternalServerError,
		Message: "internal server error; please try again later",
	}

	// return value doesn't matter; if true, then `httpErr` will be updated
	// appropriately, otherwise it will remain with the `HTTP 500` content.
	_ = errors.As(err, &httpErr)
	ctx := formData{ErrorMessage: httpErr.Message, PrivateError: err.Error()}
	return pz.Response{
		Status: httpErr.Status,
		Data:   pz.HTMLTemplate(form, &ctx),
	}.WithLogging(&ctx)
}

func parseForm(r pz.Request) (url.Values, error) {
	// Read at most 2kb data to avoid DOS attack. That should be plenty for our
	// form.
	data, err := ioutil.ReadAll(io.LimitReader(r.Body, 2056))
	if err != nil {
		return nil, formParseErr(err)
	}
	form, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, formParseErr(err)
	}
	return form, nil
}

func formParseErr(err error) *pz.HTTPError {
	return &pz.HTTPError{
		Status:  http.StatusBadRequest,
		Message: "error parsing form data",
		Cause_:  err,
	}
}

type genericFormData struct {
	Title  string
	Action html.HTMLAttr
	Fields []field
}

type formData struct {
	Token        string `json:"-"`                      // don't log the token
	ErrorMessage string `json:"errorMessage,omitempty"` // for html template
	PrivateError string `json:"privateError,omitempty"` // logging only
}

//go:embed form-template.html
var formTemplate_ string
var formTemplate = Must(template(formTemplate_))

func template(template string) (*html.Template, error) {
	return html.New("").Funcs(html.FuncMap{
		"noescape": func(s string) html.HTML { return html.HTML(s) },
	}).Parse(template)
}
