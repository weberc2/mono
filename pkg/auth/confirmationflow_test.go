package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	html "html/template"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/pkg/auth/testsupport"
	"github.com/weberc2/mono/pkg/auth/types"
)

func TestConfirmationFlow(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		flowParams   confirmationFlowParams
		users        testsupport.UserStoreFake
		request      pz.Request
		route        func(*WebServer, *confirmationFlow) pz.Route
		wantedUsers  []types.Credentials
		wantedStatus int
		wantedBody   wantedBody
	}{
		{
			name:    "main-form",
			request: pz.Request{URL: must(url.Parse("https://example.org"))},
			route:   mainFormRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields: []field{{
					ID:    "username",
					Label: "Username",
				}, {
					ID:    "password",
					Label: "Password",
					Type:  "password",
				}, {
					ID:     "token",
					Hidden: true,
					Value:  "<token>",
				}},
			},
			wantedStatus: http.StatusOK,
			wantedBody: wantedFormDocument(&formDocument{
				title:  "Activity",
				action: "/activity",
				fields: []input{{
					name:  "username",
					type_: "text",
					value: "",
					label: "Username",
				}, {
					name:  "password",
					type_: "password",
					value: "",
					label: "Password",
				}, {
					name:  "token",
					type_: "hidden",
					value: "<token>",
					label: "",
				}},
			}),
		},
		{
			name: "main-handler-simple",
			request: pz.Request{
				Body: strings.NewReader(url.Values{}.Encode()),
			},
			route: mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				mainCallback: func(
					auth *AuthService,
					form url.Values,
				) (types.UserID, error) {
					return "", nil
				},
			},
			wantedStatus: http.StatusAccepted,
			wantedBody: wantedTemplate(
				ackPageTemplate,
				&struct{ Activity string }{"Activity"},
			),
		},
		{
			name:    "main-handler-invalid-request-body",
			request: pz.Request{Body: strings.NewReader(";")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", nil
				},
			},
			wantedStatus: http.StatusBadRequest,
			wantedBody: wantedHTTPError(&pz.HTTPError{
				Status:  400,
				Message: "error parsing form data",
			}),
		},
		{
			name:    "main-handler-callback-error-500",
			request: pz.Request{Body: strings.NewReader("")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   nil,
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", fmt.Errorf("TEST ERROR")
				},
			},
			wantedStatus: http.StatusInternalServerError,
			wantedBody: wantedFormDocument(&formDocument{
				title:        "Activity",
				action:       "/activity",
				errorMessage: "internal server error",
			}),
		},
		{
			name:    "main-handler-callback-error-inherited",
			request: pz.Request{Body: strings.NewReader("")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   nil,
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", &pz.HTTPError{Status: 700, Message: "TEST ERR"}
				},
			},
			wantedStatus: 700,
			wantedBody: wantedFormDocument(&formDocument{
				title:        "Activity",
				action:       "/activity",
				errorMessage: "TEST ERR",
			}),
		},
		{
			// When the callback returns a `*successSentinelErr`, the `success`
			// callback should be invoked.
			name:    "main-handler-callback-error-sentinel",
			request: pz.Request{Body: strings.NewReader("")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   nil,
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", &successSentinelErr{
						message: "TEST ERROR",
						err:     fmt.Errorf("TEST ERROR"),
					}
				},
			},
			wantedStatus: http.StatusAccepted,
			wantedBody: wantedTemplate(
				ackPageTemplate,
				&struct{ Activity string }{"Activity"},
			),
		},
		{
			name: "confirmation-form",
			request: pz.Request{
				URL: must(url.Parse("https://example.org?t=token")),
			},
			route: confirmationFormRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
			},
			wantedStatus: http.StatusOK,
			wantedBody: wantedFormDocument(&formDocument{
				title:  "Confirm Activity",
				action: "/activity/confirm",
				fields: []input{{
					name:  "password",
					type_: "password",
					value: "",
					label: "Password",
				}, {
					name:  "token",
					type_: "hidden",
					value: "token",
					label: "",
				}},
			}),
		},
		{
			name: "confirmation-handler-simple-create",
			request: pz.Request{
				Body: strings.NewReader(url.Values{
					"token": []string{must(resetToken(
						"user",
						"user@example.org",
					))},
					"password": []string{goodPassword},
				}.Encode()),
			},
			route: confirmationHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				create:   true,
			},
			wantedUsers: []types.Credentials{{
				User:     "user",
				Email:    "user@example.org",
				Password: goodPassword,
			}},
			wantedStatus: http.StatusSeeOther,
			wantedBody:   wantedString("303 See Other"),
		},
		{
			name: "confirmation-handler-simple-no-create",
			request: pz.Request{
				Body: strings.NewReader(url.Values{
					"token": []string{must(resetToken(
						"user",
						"user@example.org",
					))},
					"password": []string{goodPassword},
				}.Encode()),
			},
			route: confirmationHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				create:   false,
			},
			users: testsupport.UserStoreFake{
				"user": &types.UserEntry{
					User:  "user",
					Email: "user@example.org",
				},
			},
			wantedStatus: http.StatusSeeOther,
			wantedBody:   wantedString("303 See Other"),
			wantedUsers: []types.Credentials{{
				User:     "user",
				Email:    "user@example.org",
				Password: goodPassword,
			}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			flow, err := newConfirmationFlow(&testCase.flowParams)
			if err != nil {
				t.Fatalf(
					"unexpected error building `confirmationFlow`: %v",
					err,
				)
			}
			if testCase.users == nil {
				testCase.users = testsupport.UserStoreFake{}
			}
			rsp := testCase.route(&WebServer{
				AuthService: AuthService{
					Creds:         CredStore{Users: testCase.users},
					ResetTokens:   resetTokenFactory,
					Tokens:        testsupport.TokenStoreFake{},
					Notifications: &testsupport.NotificationServiceFake{},
					Codes:         codesTokenFactory,
					TokenDetails: TokenDetailsFactory{
						AccessTokens:  accessTokenFactory,
						RefreshTokens: refreshTokenFactory,
						TimeFunc:      nowTimeFunc,
					},
					TimeFunc: nowTimeFunc,
				},
				BaseURL:                 "https://auth.example.org",
				RedirectDomain:          "https://app.example.org",
				DefaultRedirectLocation: "https://app.example.org",
			}, flow).Handler(testCase.request)

			if rsp.Status != testCase.wantedStatus {
				t.Fatalf(
					"wanted status `%d`; found `%d`",
					testCase.wantedStatus,
					rsp.Status,
				)
			}

			if err := testCase.wantedBody(rsp.Data); err != nil {
				t.Fatal(err)
			}

			if err := testCase.users.ExpectUsers(
				testCase.wantedUsers,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func mainFormRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeMainForm()
}

func confirmationFormRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeConfirmationForm()
}

func mainHandlerRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeMainHandler(ws)
}

func confirmationHandlerRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeConfirmationHandler(ws)
}

func wantedHTTPError(wanted *pz.HTTPError) wantedBody {
	return wantedData(func(data []byte) error {
		var found pz.HTTPError
		if err := json.Unmarshal(data, &found); err != nil {
			return fmt.Errorf(
				"response body: wanted json; found `%s`",
				data,
			)
		}
		if found.Status != wanted.Status {
			return fmt.Errorf(
				"response body: `status` field: wanted `%d`; "+
					"found `%d`",
				wanted.Status,
				found.Status,
			)
		}
		if found.Message != wanted.Message {
			return fmt.Errorf(
				"response body: `message` field: wanted `%s`; found "+
					"`%s`",
				wanted.Message,
				found.Message,
			)
		}
		return nil
	})
}

func wantedTemplate(t *html.Template, x interface{}) wantedBody {
	return wantedData(func(data []byte) error {
		var buf bytes.Buffer
		if err := t.Execute(&buf, x); err != nil {
			return fmt.Errorf("executing template: %w", err)
		}

		if wanted := buf.Bytes(); !bytes.Equal(data, wanted) {
			return fmt.Errorf(
				"unexpected response body. "+
					"wanted:\n\n%s\n\nfound:\n\n%s",
				wanted,
				data,
			)
		}

		return nil
	})
}

func wantedData(f func(data []byte) error) wantedBody {
	return func(s pz.Serializer) error {
		data, err := pztest.ReadAll(s)
		if err != nil {
			return fmt.Errorf("reading serializer: %w", err)
		}
		return f(data)
	}
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

type formDocument struct {
	title        string
	action       string
	errorMessage string
	fields       []input
}

func wantedFormDocument(wanted *formDocument) wantedBody {
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

type input struct {
	name  string
	type_ string
	value string
	label string
}

type wantedBody func(pz.Serializer) error
