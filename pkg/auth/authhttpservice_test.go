package auth

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/pkg/auth/testsupport"
	"github.com/weberc2/mono/pkg/auth/types"
	. "github.com/weberc2/mono/pkg/prelude"
)

func TestAuthHTTPService(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		input          string
		route          func(*AuthHTTPService) pz.Route
		validationTime time.Time
		existingTokens testsupport.TokenStoreFake
		existingUsers  []types.UserEntry
		wantedStatus   int
		wantedPayload  pztest.WantedData
		wantedTokens   []types.Token
	}{
		{
			name: "forgot password",
			existingUsers: []types.UserEntry{{
				User:         "user",
				Email:        "user@example.org",
				PasswordHash: Must(testsupport.HashBcrypt("password")),
			}},
			existingTokens: testsupport.TokenStoreFake{},
			input:          `{"user": "user"}`,
			route:          (*AuthHTTPService).ForgotPasswordRoute,
			wantedStatus:   200,
			wantedPayload:  Any{},
		},
		{
			// Still want to return 200 when user isn't found to avoid leaking
			// details to potential attackers.
			name:           "forgot password: user not found",
			existingUsers:  nil,
			existingTokens: testsupport.TokenStoreFake{},
			input:          `{"user": "user"}`,
			route:          (*AuthHTTPService).ForgotPasswordRoute,
			wantedStatus:   200,
			wantedPayload:  Any{},
		},
		{
			// Expect tokens are returned when a valid refresh token is
			// provided.
			name: "refresh",
			existingTokens: testsupport.TokenStoreFake{
				testsupport.RefreshToken.Token: testsupport.RefreshToken.
					Expires,
			},
			input: fmt.Sprintf(
				`{"refreshToken": "%s"}`,
				testsupport.RefreshToken.Token,
			),
			route:          (*AuthHTTPService).RefreshRoute,
			validationTime: testsupport.Now.Add(2 * time.Second),
			wantedStatus:   200,
			wantedPayload: &RefreshResponse{
				AccessToken: testsupport.AccessToken.Token,
			},
			wantedTokens: []types.Token{testsupport.RefreshToken},
		},
		{
			// Expect an error when an invalid refresh token is provided. The
			// same generic `invalid token` error is used regardless of the
			// nature of the error to avoid leaking information to potential
			// attackers.
			name:           "refresh: invalid token",
			existingTokens: testsupport.TokenStoreFake{},
			input:          `{"refreshToken": "foobar"}`,
			route:          (*AuthHTTPService).RefreshRoute,
			validationTime: testsupport.Now.Add(2 * time.Second),
			wantedStatus:   401,
			wantedPayload:  ErrInvalidRefreshToken,
		},
		{
			// Expect an error when an expired refresh token is provided. The
			// same generic `invalid token` error is used regardless of the
			// nature of the error to avoid leaking information to potential
			// attackers.
			name:           "refresh: expired token",
			existingTokens: testsupport.TokenStoreFake{},
			input: fmt.Sprintf(
				`{"refreshToken": "%s"}`,
				testsupport.RefreshToken.Token,
			),
			route:          (*AuthHTTPService).RefreshRoute,
			validationTime: testsupport.Now.Add(30 * 24 * time.Hour),
			wantedStatus:   401,
			wantedPayload:  ErrInvalidRefreshToken,
		},
		{
			// Expect ErrTokenNotFound when an unknown refresh token is
			// provided.
			name:           "refresh: unknown token",
			existingTokens: testsupport.TokenStoreFake{},
			input: fmt.Sprintf(
				`{"refreshToken": "%s"}`,
				testsupport.RefreshToken.Token,
			),
			route:          (*AuthHTTPService).RefreshRoute,
			validationTime: testsupport.Now.Add(2 * time.Second),
			wantedStatus:   401,
			wantedPayload:  types.ErrTokenNotFound,
		},
		{
			// Expect tokens returned in exchange for a valid auth code.
			name: "exchange auth code",
			input: fmt.Sprintf(
				`{"code": "%s"}`,
				testsupport.AuthCode.Token,
			),
			route:          (*AuthHTTPService).ExchangeRoute,
			existingTokens: testsupport.TokenStoreFake{},
			validationTime: testsupport.Now,
			wantedStatus:   200,
			wantedPayload: &TokenDetails{
				AccessToken:  testsupport.AccessToken,
				RefreshToken: testsupport.RefreshToken,
			},
			wantedTokens: []types.Token{testsupport.RefreshToken},
		},
		{
			name: "logout",
			existingTokens: testsupport.TokenStoreFake{
				testsupport.RefreshToken.Token: testsupport.RefreshToken.
					Expires,
			},
			route: (*AuthHTTPService).LogoutRoute,
			input: fmt.Sprintf(
				`{"refreshToken": "%s"}`,
				testsupport.RefreshToken.Token,
			),
			wantedStatus: 200,
			wantedPayload: &LogoutResponse{
				Status:  200,
				Message: "successfully logged out",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			vtime := testsupport.Now
			if testCase.validationTime != (time.Time{}) {
				vtime = testCase.validationTime
			}
			jwt.TimeFunc = func() time.Time { return vtime }
			defer func() { jwt.TimeFunc = time.Now }()
			var notifications []*types.Notification
			service := AuthHTTPService{
				AuthService: AuthService{
					Tokens: testCase.existingTokens,
					Creds: CredStore{
						Users: &userStoreMock{
							get: func(
								user types.UserID,
							) (*types.UserEntry, error) {
								for i, entry := range testCase.existingUsers {
									if entry.User == user {
										return &testCase.existingUsers[i], nil
									}
								}
								return nil, types.ErrUserNotFound
							},
						},
					},
					Notifications: &notificationServiceMock{
						notify: func(n *types.Notification) error {
							notifications = append(notifications, n)
							return nil
						},
					},
					Codes:       testsupport.CodesTokenFactory,
					ResetTokens: testsupport.ResetTokenFactory,
					TokenDetails: TokenDetailsFactory{
						AccessTokens:  testsupport.AccessTokenFactory,
						RefreshTokens: testsupport.RefreshTokenFactory,
						TimeFunc:      testsupport.NowTimeFunc,
					},
					TimeFunc: testsupport.NowTimeFunc,
				},
			}

			rsp := testCase.route(&service).Handler(pz.Request{
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

			if err := pztest.CompareSerializer(
				testCase.wantedPayload,
				rsp.Data,
			); err != nil {
				if err, ok := err.(*pztest.CompareSerializerError); ok {
					t.Logf("response data: %s", err.Data)
				}
				t.Fatal(err)
			}

			found, _ := testCase.existingTokens.List()
			if err := compareManyTokens(
				&testsupport.RefreshTokenFactory.SigningKey.PublicKey,
				testCase.wantedTokens,
				found,
			); err != nil {
				t.Fatalf("checking token store: %v", err)
			}
		})
	}
}

func compareManyTokens(
	key *ecdsa.PublicKey,
	wanted []types.Token,
	found []types.Token,
) error {
	if len(wanted) != len(found) {
		return fmt.Errorf(
			"wanted %d tokens; found %d tokens",
			len(wanted),
			len(found),
		)
	}
	for i := range wanted {
		if err := compareTokens(
			key,
			wanted[i].Token,
			found[i].Token,
		); err != nil {
			return fmt.Errorf("token %d: %w", i, err)
		}
		if !wanted[i].Expires.Equal(found[i].Expires) {
			return fmt.Errorf(
				"token %d: Token.Expires: wanted `%s`; found `%s`",
				i,
				wanted[i].Expires,
				found[i].Expires,
			)
		}
	}

	return nil
}

func compareTokens(key *ecdsa.PublicKey, wanted, found string) error {
	var wantedClaims types.Claims
	if _, err := jwt.ParseWithClaims(
		wanted,
		&wantedClaims,
		func(*jwt.Token) (interface{}, error) { return key, nil },
	); err != nil {
		return fmt.Errorf("parsing 'wanted' token: %w", err)
	}

	var foundClaims types.Claims
	if _, err := jwt.ParseWithClaims(
		found,
		&foundClaims,
		func(*jwt.Token) (interface{}, error) { return key, nil },
	); err != nil {
		return fmt.Errorf("parsing 'found' token: %w", err)
	}

	if wantedClaims != foundClaims {
		wanted, err := json.Marshal(wantedClaims)
		if err != nil {
			return fmt.Errorf("marshaling 'wanted' claims: %w", err)
		}
		found, err := json.Marshal(foundClaims)
		if err != nil {
			return fmt.Errorf("marshaling 'found' claims: %w", err)
		}
		return fmt.Errorf("wanted `%s`; found `%s`", wanted, found)
	}

	return nil
}

func (wanted *RefreshResponse) CompareData(data []byte) error {
	var found RefreshResponse
	if err := json.Unmarshal(data, &found); err != nil {
		return fmt.Errorf("unmarshaling `refresh`: %w", err)
	}

	if err := compareTokens(
		&testsupport.AccessSigningKey.PublicKey,
		wanted.AccessToken,
		found.AccessToken,
	); err != nil {
		return fmt.Errorf("comparing access tokens: %w", err)
	}

	return nil
}

func (wanted *TokenDetails) CompareData(data []byte) error {
	var found TokenDetails
	if err := json.Unmarshal(data, &found); err != nil {
		return fmt.Errorf(
			"TokenDetails.Compare(): unmarshaling `TokenDetails`: %w",
			err,
		)
	}

	if err := compareTokens(
		&testsupport.AccessSigningKey.PublicKey,
		wanted.AccessToken.Token,
		found.AccessToken.Token,
	); err != nil {
		log.Printf(
			"access token: wanted `%s`; found `%s`",
			wanted.AccessToken,
			found.AccessToken,
		)
		return fmt.Errorf("TokenDetails.Compare(): AccessToken: %w", err)
	}

	if err := compareTokens(
		&testsupport.RefreshSigningKey.PublicKey,
		wanted.RefreshToken.Token,
		found.RefreshToken.Token,
	); err != nil {
		log.Printf(
			"refresh token: wanted `%s`; found `%s`",
			wanted.RefreshToken,
			found.RefreshToken,
		)
		return fmt.Errorf("TokenDetails.Compare(): RefreshToken: %w", err)
	}
	return nil
}

type Any struct{}

func (Any) CompareData(data []byte) error { return nil }
