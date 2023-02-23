package auth

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/mod/auth/pkg/auth/testsupport"
)

type formDocument struct {
	title        string
	action       string
	errorMessage string
	fields       []input
}

func wantedFormDocument(wanted *formDocument) testsupport.WantedBody {
	return func(data pz.Serializer) error {
		return expectFormDocumentFromSerializer(data, wanted)
	}
}

func expectFormDocumentFromSerializer(
	s pz.Serializer,
	wanted *formDocument,
) error {
	data, err := pztest.ReadAll(s)
	if err != nil {
		return fmt.Errorf("reading serializer: %w", err)
	}

	d, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("HTML-parsing response: %w", err)
	}
	return expectFormDocument(d, wanted)

}

func expectFormDocument(d *goquery.Document, wanted *formDocument) error {
	head := d.Find("head")
	if head.Length() < 1 {
		return newError(d.Selection, "form document: missing `<head>` element")
	}

	titleNode := head.Find("title")
	if titleNode.Length() < 1 {
		return newError(
			head,
			"form document: `<head>` element: missing `<title>` element",
		)
	}

	if titleText := titleNode.Text(); titleText != wanted.title {
		return newError(
			titleNode,
			"form document: `<head>` element: `<title> element: wanted "+
				"`%s`; found `%s`",
			wanted.title,
			titleText,
		)
	}

	body := d.Find("body")
	if body.Length() < 1 {
		return newError(d.Selection, "form document: missing `<body>` element")
	}
	titleH1 := body.Find("h1#title")
	if titleH1.Length() < 1 {
		return newError(
			body,
			"form document: `<body>` element: missing "+
				"`<h1 id=\"title\">` element",
		)
	}
	if titleText := titleH1.Text(); titleText != wanted.title {
		return newError(
			titleH1,
			"form document: `<body>` element: `<h1 id=\"title\">` element: "+
				"wanted `%s`; found `%s`",
			wanted.title,
			titleText,
		)
	}

	errorMessage := body.Find("p#error-message")
	if wanted.errorMessage != "" {
		if errorMessage.Length() < 1 {
			return newError(
				body,
				"form document: `<body>` element: missing "+
					"`<p id=\"error-message>` element",
			)
		}
		if foundErr := errorMessage.Text(); foundErr != wanted.errorMessage {
			return newError(
				body,
				"form document: `<body>` element: `<p id=\"error-message\">` "+
					"element: wanted `%s`; found `%s`",
				wanted.errorMessage,
				foundErr,
			)
		}
	} else if errorMessage.Length() > 0 {
		return newError(
			body,
			"form document: `<body>` element: unexpected element: "+
				"`<p id=\"error-message\">`",
		)
	}

	return expectForm(d, wanted.action, wanted.fields)
}

func expectForm(
	d *goquery.Document,
	action string,
	fields []input,
) error {
	form := d.Find("form")
	if form.Length() < 1 {
		return newError(
			d.Selection,
			"document has no `<form>` element",
		)
	}

	foundAction, exists := form.First().Attr("action")
	if !exists {
		return newError(form, "form missing attribute: `action`")
	}

	if foundAction != action {
		return fmt.Errorf(
			"form action: wanted `%s`; found `%s`",
			action,
			foundAction,
		)
	}

	for i := range fields {
		if err := expectInput(form, &fields[i]); err != nil {
			return err
		}
	}

	return nil
}

func expectInput(
	form *goquery.Selection,
	input *input,
) error {
	field := form.Find(fmt.Sprintf("input[name=\"%s\"]", input.name))
	if field.Length() < 1 {
		return newError(form, "form field `%s`: not found", input.name)
	}

	type_, exists := field.First().Attr("type")
	if !exists {
		return newError(
			form,
			"form field `%s`: missing `type` attribute",
			input.name,
		)
	}

	if type_ != input.type_ {
		return newError(
			field,
			"form's field `%s`: attribute `type`: wanted `%s`; found `%s`",
			input.name,
			input.type_,
			type_,
		)
	}

	value, exists := field.First().Attr("value")
	if input.value == "" && exists {
		return newError(
			field,
			"form field `%s`: wanted no `value` attribute; found `%s`",
			input.name,
			value,
		)
	} else if input.value != "" && !exists {
		return newError(
			field,
			"form field `%s`: attribute `value`: wanted `%s` but attribute "+
				"not found",
			input.name,
			input.value,
		)
	} else if input.value != "" && value != input.value {
		return newError(
			field,
			"form field `%s`: attribute `value`: wanted `%s`; found `%s`",
			input.name,
			input.value,
			value,
		)
	}

	label := form.Find(fmt.Sprintf(`label[for="%s"]`, input.name))
	if input.label == "" && label.Length() > 0 {
		return newError(
			form,
			"form field `%s`: found unwanted label",
			input.name,
		)
	} else if input.label != "" && label.Length() < 1 {
		return newError(
			form,
			"form field `%s`: wanted label `%s`; label node not found",
			input.name,
			input.label,
		)
	} else if input.label != "" && label.Text() != input.label {
		return newError(
			label,
			"form field `%s`: wanted label `%s`; found `%s`",
			input.name,
			input.label,
			label.Text(),
		)
	}

	return nil
}

func newError(s *goquery.Selection, format string, v ...interface{}) error {
	return handleErr(s, fmt.Errorf(format, v...))
}

func handleErr(s *goquery.Selection, err error) error {
	html, renderErr := s.Html()
	if renderErr != nil {
		return fmt.Errorf(
			"error rendering HTML while handling error (original error: "+
				"%w): %v",
			err,
			renderErr,
		)
	}

	return fmt.Errorf("%w:\n\n%s", err, html)
}

type input struct {
	name  string
	type_ string
	value string
	label string
}

type kv [2]string

func encodeForm(form ...kv) string {
	const key = 0
	const val = 1
	if len(form) < 1 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(url.QueryEscape(form[0][key]))
	sb.WriteByte('=')
	sb.WriteString(url.QueryEscape(form[0][val]))
	for i := range form[1:] {
		sb.WriteByte('&')
		sb.WriteString(url.QueryEscape(form[i+1][key]))
		sb.WriteByte('=')
		sb.WriteString(url.QueryEscape(form[i+1][val]))
	}
	return sb.String()
}
