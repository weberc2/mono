package main

import (
	"errors"

	pz "github.com/weberc2/httpeasy"
)

type AuthHTTPService struct {
	AuthService
}

func (ahs *AuthHTTPService) LoginRoute() pz.Route {
	return pz.Route{
		Path:   "/api/login",
		Method: "POST",
		Handler: func(r pz.Request) pz.Response {
			var creds Credentials
			if err := r.JSON(&creds); err != nil {
				return pz.BadRequest(nil, struct {
					Message, Error string
				}{
					Message: "failed to parse login JSON",
					Error:   err.Error(),
				})
			}

			tokens, err := ahs.Login(&creds)
			if err != nil {
				if errors.Is(err, ErrCredentials) {
					return pz.Unauthorized(
						pz.String("Invalid username or password"),
						struct {
							Message string
							User    UserID
						}{
							Message: "invalid username or password",
							User:    creds.User,
						},
					)
				}

				return pz.InternalServerError(
					struct {
						Message, Error string
						User           UserID
					}{
						Message: "logging in",
						Error:   err.Error(),
						User:    creds.User,
					},
				)
			}

			return pz.Ok(
				pz.JSON(tokens),
				struct {
					Message string
					User    UserID
				}{
					Message: "authentication succeeded",
					User:    creds.User,
				},
			)
		},
	}
}

func (ahs *AuthHTTPService) RefreshRoute() pz.Route {
	return pz.Route{
		Path:   "/api/refresh",
		Method: "POST",
		Handler: func(r pz.Request) pz.Response {
			var payload struct {
				RefreshToken string `json:"refreshToken"`
			}
			if err := r.JSON(&payload); err != nil {
				return pz.BadRequest(nil, struct{ Message, Error string }{
					Message: "failed to parse refresh JSON",
					Error:   err.Error(),
				})
			}

			accessToken, err := ahs.Refresh(payload.RefreshToken)
			if err != nil {
				return pz.InternalServerError(struct{ Message, Error string }{
					Message: "refreshing access token",
					Error:   err.Error(),
				})
			}

			return pz.Ok(pz.JSON(struct {
				AccessToken string `json:"accessToken"`
			}{accessToken}))
		},
	}
}

func (ahs *AuthHTTPService) ForgotPasswordRoute() pz.Route {
	return pz.Route{
		Path:   "/api/password/forgot",
		Method: "POST",
		Handler: func(r pz.Request) pz.Response {
			var payload struct {
				User UserID `json:"user"`
			}
			if err := r.JSON(&payload); err != nil {
				return pz.BadRequest(nil, struct{ Message, Error string }{
					Message: "failed to parse forgot-password JSON",
					Error:   err.Error(),
				})
			}

			if err := ahs.ForgotPassword(payload.User); err != nil {
				// If the user doesn't exist, we still report success so as to
				// not give away information to potential attackers.
				if errors.Is(err, ErrUserNotFound) {
					return pz.Ok(nil, struct{ Message, User, Error string }{
						Message: "user not found; silently succeeding",
						User:    string(payload.User),
						Error:   err.Error(),
					})
				}

				return pz.InternalServerError(struct {
					Message, User, Error string
				}{
					Message: "triggering forget-password notification",
					User:    string(payload.User),
					Error:   err.Error(),
				})
			}

			return pz.Ok(nil, struct{ Message, User string }{
				Message: "password reset notification sent",
				User:    string(payload.User),
			})
		},
	}
}

func (ahs *AuthHTTPService) RegisterRoute() pz.Route {
	return pz.Route{
		Path:   "/api/register",
		Method: "POST",
		Handler: func(r pz.Request) pz.Response {
			var payload struct {
				User  UserID `json:"user"`
				Email string `json:"email"`
			}
			if err := r.JSON(&payload); err != nil {
				return pz.BadRequest(nil, struct{ Message, Error string }{
					Message: "failed to parse register JSON",
					Error:   err.Error(),
				})
			}

			if err := ahs.Register(payload.User, payload.Email); err != nil {
				if errors.Is(err, ErrInvalidEmail) {
					return pz.BadRequest(
						pz.String("Invalid email address"),
						struct {
							Error string
						}{
							Error: err.Error(),
						},
					)
				}
				if errors.Is(err, ErrUserExists) {
					return pz.Conflict(
						pz.String("User already exists"),
						struct {
							Message, Error string
							User           UserID
						}{
							Message: "registering user",
							Error:   err.Error(),
							User:    payload.User,
						},
					)
				}
				return pz.InternalServerError(struct {
					Message, Error string
					User           UserID
				}{
					Message: "registering user",
					Error:   err.Error(),
					User:    payload.User,
				})
			}

			return pz.Created(pz.String("Created user"), struct {
				Message string
				User    UserID
			}{
				Message: "created user",
				User:    payload.User,
			})
		},
	}
}

func (ahs *AuthHTTPService) UpdatePasswordRoute() pz.Route {
	return pz.Route{
		Path:   "/api/password",
		Method: "PATCH",
		Handler: func(r pz.Request) pz.Response {
			var payload UpdatePassword
			if err := r.JSON(&payload); err != nil {
				return pz.BadRequest(nil, struct {
					Message, Error string
					User           UserID
				}{
					Message: "updating password",
					Error:   err.Error(),
					User:    payload.User,
				})
			}

			if err := ahs.UpdatePassword(&payload); err != nil {
				l := struct {
					Message, Error string
					User           UserID
				}{
					Message: "updating password",
					Error:   err.Error(),
					User:    payload.User,
				}
				if errors.Is(err, ErrInvalidResetToken) {
					return pz.NotFound(
						pz.String(ErrInvalidResetToken.Error()),
						l,
					)
				}
				return pz.InternalServerError(l)
			}

			return pz.Ok(pz.String("Password updated"), struct {
				Message string
				User    UserID
			}{
				Message: "updated password",
				User:    payload.User,
			})
		},
	}
}

func (ahs *AuthHTTPService) Routes() []pz.Route {
	return []pz.Route{
		ahs.LoginRoute(),
		ahs.RefreshRoute(),
		ahs.RegisterRoute(),
		ahs.ForgotPasswordRoute(),
		ahs.UpdatePasswordRoute(),
	}
}
