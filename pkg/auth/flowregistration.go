package auth

import (
	"net/url"

	"github.com/weberc2/mono/pkg/auth/types"
)

var (
	flowRegistration = confirmationFlow{
		activity: "registration",
		main: form{
			path: "/registration",
			template: mustHTML(
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
</html>`),
			callback: func(
				auth *AuthService,
				f url.Values,
			) (types.UserID, error) {
				user := types.UserID(f.Get("username"))
				return user, auth.Register(user, f.Get("email"))
			},
			success: successAccepted(registrationAckPage),
		},
		confirmation: form{
			path: "/registration/confirmation",
			template: mustHTML(`<html>
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
</html>`),
			callback: callbackUpdatePassword(true),
			success:  successDefaultRedirect,
		},
	}
)

const registrationAckPage = `<html>
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
