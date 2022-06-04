package auth

import (
	"net/url"

	"github.com/weberc2/mono/pkg/auth/types"
)

var (
	flowPasswordReset = confirmationFlow{
		activity: "password reset",
		main: form{
			path: "/password-reset",
			template: mustHTML(`<html>
<head>
	<title>Password Reset</title>
</head>
<body>
<h1>Password Reset</h1>
{{ if .ErrorMessage }}<p id="error-message">{{ .ErrorMessage }}</p>{{ end }}
<form action="{{ .FormAction }}" method="POST">
	<label for="username">Username</label>
	<input type="text" id="username" name="username"><br><br>
	<input type="submit" value="Submit">
</form>
</body>
</html>`),
			callback: func(
				auth *AuthService,
				f url.Values,
			) (types.UserID, error) {
				user := types.UserID(f.Get("username"))
				return user, auth.ForgotPassword(user)
			},
			success: successAccepted(`<html>
<head>
	<title>Initiated Password Reset</title>
<body>
<h1>Initiated Password Reset</h1>
<p>An email has been sent to the email address corresponding to the provided
username. Please check your email for a confirmation link.</p>
</body>
</head>
</html>`),
		},
		confirmation: form{
			path: "/password-reset/confirmation",
			template: mustHTML(`<html>
<head>
	<title>Confirm Password Reset</title>
</head>
<body>
<h1>Confirm Password Reset</h1>
{{ if .ErrorMessage }}<p id="error-message">{{ .ErrorMessage }}</p>{{ end }}
<form action="{{ .FormAction }}" method="POST">
	<label for="password">Password</label>
	<input type="password" id="password" name="password"><br><br>
	<input type="hidden" id="token" name="token" value="{{.Token}}">
	<input type="submit" value="Submit">
</form>
</body>
</html>`),
			callback: callbackUpdatePassword(false),
			success:  successDefaultRedirect,
		},
	}
)
