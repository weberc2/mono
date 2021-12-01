package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/weberc2/auth/pkg/auth"
	pz "github.com/weberc2/httpeasy"
)

type WebServerApp struct {
	Client          Client
	BaseURL         string
	DefaultRedirect string
	Key             string
}

func (app *WebServerApp) DecryptAccessToken(r pz.Request) (string, error) {
	return app.DecryptCookie(r, "Access-Token")
}

func (app *WebServerApp) DecryptRefreshToken(r pz.Request) (string, error) {
	return app.DecryptCookie(r, "Refresh-Token")
}

func (app *WebServerApp) DecryptCookie(
	r pz.Request,
	cookie string,
) (string, error) {
	c, err := r.Cookie(cookie)
	if err != nil {
		return "", fmt.Errorf("finding cookie `%s`: %w", cookie, err)
	}
	plaintext, err := app.decryptCookie(c)
	if err != nil {
		return "", fmt.Errorf("decrypting `%s` cookie: %w", cookie, err)
	}
	return plaintext, nil
}

func (app *WebServerApp) decryptCookie(cookie *http.Cookie) (string, error) {
	plaintext, err := decrypt(cookie.Value, app.Key)
	if err != nil {
		return "", fmt.Errorf("decrypting cookie: %w", err)
	}
	return plaintext, nil
}

func (app *WebServerApp) Encrypt(data string) (string, error) {
	return encrypt(data, app.Key)
}

func join(base, path string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(base, "/"),
		strings.TrimLeft(path, "/"),
	)
}

func (app *WebServerApp) AuthCodeCallbackRoute(path string) pz.Route {
	return pz.Route{
		Path:   path,
		Method: "GET",
		Handler: func(r pz.Request) pz.Response {
			query := r.URL.Query()
			context := struct {
				Message            string `json:"message,omitempty"`
				RedirectSpecified  string `json:"redirectSpecified"`
				RedirectActual     string `json:"redirectActual"`
				RedirectParseError string `json:"redirectParseError,omitempty"`
				CodeParseError     string `json:"codeParseError,omitempty"`
				EncryptionError    string `json:"encryptionError,omitempty"`
				BaseURLParseError  string `json:"baseURLParseError,omitempty"`
			}{
				RedirectSpecified: join(app.BaseURL, query.Get("redirect")),
			}
			if context.RedirectSpecified == "" {
				context.RedirectParseError = "`redirect` query param " +
					"missing or empty"
				context.RedirectActual = app.DefaultRedirect
			} else if _, err := url.Parse(
				context.RedirectSpecified,
			); err != nil {
				context.RedirectParseError = err.Error()
				context.RedirectActual = join(app.BaseURL, app.DefaultRedirect)
			} else {
				context.RedirectActual = context.RedirectSpecified
			}

			code := query.Get("code")
			if code == "" {
				context.CodeParseError = "`code` query string parameter " +
					"missing or empty"
				return pz.BadRequest(nil, &context)
			}

			tokens, err := app.Client.Exchange(code)
			if err != nil {
				return pz.HandleError("exchanging auth code", err)
			}

			if err := encryptTokens(tokens, app.Key); err != nil {
				context.EncryptionError = err.Error()
				return pz.InternalServerError(&context)
			}

			u, err := url.Parse(app.BaseURL)
			if err != nil {
				context.BaseURLParseError = err.Error()
				return pz.InternalServerError(&context)
			}

			portStart := strings.Index(u.Host, ":")
			if portStart < 0 {
				portStart = len(u.Host)
			}

			context.Message = "successfully exchanged auth code"
			return pz.SeeOther(context.RedirectActual, &context).WithCookies(
				&http.Cookie{
					Name:     "Access-Token",
					Domain:   u.Host[:portStart],
					Value:    tokens.AccessToken,
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				},
				&http.Cookie{
					Name:     "Refresh-Token",
					Domain:   u.Host[:portStart],
					Value:    tokens.RefreshToken,
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				},
			)
		},
	}
}

func encryptTokens(tokens *auth.TokenDetails, key string) error {
	access, err := encrypt(tokens.AccessToken, key)
	if err != nil {
		return fmt.Errorf("encrypting access token: %w", err)
	}
	refresh, err := encrypt(tokens.RefreshToken, key)
	if err != nil {
		return fmt.Errorf("encrypting refresh token: %w", err)
	}
	tokens.AccessToken, tokens.RefreshToken = access, refresh
	return nil
}

func encrypt(input, key string) (string, error) {
	plaintext := []byte(input)

	k := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("populating nonce: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(gcm.Seal(
		nonce,
		nonce,
		plaintext,
		nil,
	)), nil
}

func decrypt(input, key string) (string, error) {
	encrypted := []byte(input)
	k := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(k[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nsz := gcm.NonceSize()
	data, err := gcm.Open(nil, encrypted[:nsz], encrypted[nsz:], nil)
	return string(data), err
}
