package auth

import (
	"bytes"
	"fmt"
	html "html/template"
	"strings"
	text "text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/weberc2/mono/mod/comments/pkg/auth/types"
)

type NotificationSettings struct {
	HTMLTemplate *html.Template
	TextTemplate *text.Template
	TokenURL     string
	Subject      string
}

func DefaultRegistrationSettings(baseURL string) NotificationSettings {
	return NotificationSettings{
		Subject: "Register account",
		HTMLTemplate: html.Must(html.New("").Parse(`<p>Hello,<br /><br />

Someone has attempted to register an account with this email address. If this was not you, please disregard this message. If this was intentional, please click this <a href="{{ .TokenURL }}">link</a> to finish creating your account.</p>`)),
		TextTemplate: text.Must(text.New("").Parse(`Hello,

Someone has attempted to register an account with this email address. If this was not you, please disregard this message. If this was intentional, please enter the following URL into your web browser to finish creating your account: {{ .TokenURL }}`)),
		TokenURL: strings.TrimRight(baseURL, "/") +
			flowRegistration.confirmation.path,
	}
}

func DefaultForgotPasswordSettings(baseURL string) NotificationSettings {
	return NotificationSettings{
		Subject: "Forgot password",
		HTMLTemplate: html.Must(
			html.New("").Parse(`<p>Hello {{ .User }},<br /><br />

Someone has attempted to reset your password. If this was not you, please disregard this message. If this was intentional, please click this <a href="{{ .TokenURL }}">link</a> to reset your password.</p>`),
		),
		TextTemplate: text.Must(text.New("").Parse(`Hello {{ .User }},
Someone has attempted to reset your password. If this was not you, please disregard this message. If this was intentional, please enter the following URL into your web browser to reset your password: {{ .TokenURL }}`)),
		TokenURL: strings.TrimRight(baseURL, "/") +
			flowPasswordReset.confirmation.path,
	}
}

type SESNotificationService struct {
	Client                 *ses.SES
	Sender                 string
	RegistrationSettings   NotificationSettings
	ForgotPasswordSettings NotificationSettings
}

func (sns *SESNotificationService) Notify(token *types.Notification) error {
	var settings *NotificationSettings
	if token.Type == types.NotificationTypeRegister {
		settings = &sns.RegistrationSettings
	} else {
		settings = &sns.ForgotPasswordSettings
	}

	var htmlBuf, textBuf bytes.Buffer
	payload := struct {
		User     types.UserID
		Email    string
		TokenURL string
	}{
		User:     token.User,
		Email:    token.Email,
		TokenURL: fmt.Sprintf("%s?t=%s", settings.TokenURL, token.Token),
	}

	if err := settings.TextTemplate.Execute(&htmlBuf, &payload); err != nil {
		return fmt.Errorf("rendering text template: %w", err)
	}
	if err := settings.HTMLTemplate.Execute(&textBuf, &payload); err != nil {
		return fmt.Errorf("rendering html template: %w", err)
	}
	if _, err := sns.Client.SendEmail(&ses.SendEmailInput{
		Destination: &ses.Destination{ToAddresses: []*string{&token.Email}},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    &settings.Subject,
			},
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(htmlBuf.String()),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(textBuf.String()),
				},
			},
		},
		Source: &sns.Sender,
	}); err != nil {
		return fmt.Errorf("sending email via AWS SES: %w", err)
	}
	return nil
}
