package auth

import (
	_ "embed"
	"fmt"
	pz "github.com/weberc2/httpeasy"
	html "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type field struct {
	ID     string
	Label  string
	Type   string
	Hidden bool
	Value  string
}

func formHTML(title, action string, fields ...field) (*html.Template, error) {
	fmt.Printf("%v", fields)
	var sb strings.Builder
	if err := formTemplate.Execute(
		&sb,
		&struct {
			Title  string
			Action string
			Fields []field
		}{
			Title:  title,
			Action: action,
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

func formHandler(next func(url.Values) pz.Response) pz.Handler {
	return func(r pz.Request) pz.Response {
		return handleForm(r, next)
	}
}

func handleForm(r pz.Request, next func(url.Values) pz.Response) pz.Response {
	form, err := parseForm(r)
	if err != nil {
		return handleError("error parsing form data", "parsing form data", err)
	}
	return next(form)
}

func handleError(publicMessage, privateMessage string, err error) pz.Response {
	return pz.HandleError(
		publicMessage,
		err,
		&logging{
			Message:   privateMessage,
			ErrorType: reflect.TypeOf(err).String(),
			Error:     err.Error(),
		},
	)
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

//go:embed form-template.html
var formTemplate_ string
var formTemplate = must(template(formTemplate_))
