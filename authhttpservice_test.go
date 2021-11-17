package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
)

func TestAuthHTTPService(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	signingKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	for _, testCase := range []struct {
		name          string
		input         string
		existingUsers []UserEntry
		wantedStatus  int
	}{
		{
			name:  "forgot password",
			input: `{"user": "user"}`,
			existingUsers: []UserEntry{{
				User:         "user",
				Email:        "user@example.org",
				PasswordHash: hashBcrypt("password"),
			}},
			wantedStatus: 200,
		},
		{
			// Still want to return 200 when user isn't found to avoid leaking
			// details to potential attackers.
			name:          "forgot password: user not found",
			input:         `{"user": "user"}`,
			existingUsers: nil,
			wantedStatus:  200,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var notifications []*Notification
			service := AuthHTTPService{
				AuthService: AuthService{
					Creds: CredStore{
						Users: &userStoreMock{
							get: func(user UserID) (*UserEntry, error) {
								for i, entry := range testCase.existingUsers {
									if entry.User == user {
										return &testCase.existingUsers[i], nil
									}
								}
								return nil, ErrUserNotFound
							},
						},
					},
					Notifications: &notificationServiceMock{
						notify: func(n *Notification) error {
							notifications = append(notifications, n)
							return nil
						},
					},
					ResetTokens: ResetTokenFactory{
						Issuer:           "issuer",
						WildcardAudience: "audience",
						TokenValidity:    1 * time.Hour,
						SigningKey:       signingKey,
						SigningMethod:    jwt.SigningMethodES512,
					},
					TimeFunc: func() time.Time { return now },
				},
			}

			rsp := service.ForgotPasswordRoute().Handler(pz.Request{
				Body: strings.NewReader(testCase.input),
			})

			if rsp.Status != testCase.wantedStatus {
				data, err := json.Marshal(rsp.Logging)
				if err != nil {
					t.Logf("failed to marshal handler logs: %v", err)
				}
				t.Logf("request logs: %s", data)
				t.Fatalf(
					"wanted `%d`; found `%d`",
					testCase.wantedStatus,
					rsp.Status,
				)
			}
		})
	}
}
