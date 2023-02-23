package auth

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/mod/comments/pkg/auth/testsupport"
	"github.com/weberc2/mono/mod/comments/pkg/auth/types"
)

type routeTestCase struct {
	name                string
	users               testsupport.UserStoreFake
	request             pz.Request
	route               func(ws *WebServer) pz.Handler
	wantedResponse      response
	wantedUsers         []types.Credentials
	wantedNotifications []*types.Notification
}

func (rtc *routeTestCase) run() error {
	jwt.TimeFunc = testsupport.NowTimeFunc
	defer func() { jwt.TimeFunc = time.Now }()
	if rtc.users == nil {
		rtc.users = testsupport.UserStoreFake{}
	}

	var notifications testsupport.NotificationServiceFake
	if rtc.request.URL == nil {
		rtc.request.URL = new(url.URL)
	}
	if rtc.request.Headers == nil {
		rtc.request.Headers = make(http.Header)
	}
	if rtc.route == nil {
		panic(fmt.Sprintf(
			"test case `%s` left its `route` field nil",
			rtc.name,
		))
	}
	webServer := testWebServer(rtc.users, &notifications)
	if err := rtc.wantedResponse.compare(
		rtc.route(webServer)(rtc.request),
	); err != nil {
		return err
	}

	if err := rtc.users.ExpectUsers(rtc.wantedUsers); err != nil {
		return err
	}

	if err := compareNotifications(
		testsupport.ResetTokenFactory.SigningKey.PublicKey,
		rtc.wantedNotifications,
		notifications.Notifications,
	); err != nil {
		return err
	}
	return nil
}

func compareNotifications(
	key ecdsa.PublicKey,
	wanted []*types.Notification,
	found []*types.Notification,
) error {
	if len(wanted) != len(found) {
		return fmt.Errorf(
			"sent notifications: length: wanted `%d`; found `%d`",
			len(wanted),
			len(found),
		)
	}

	for i := range wanted {
		if wanted[i].Email != found[i].Email {
			return fmt.Errorf(
				"Notifications[%d].Email: wanted `%s`; found `%s`",
				i,
				wanted[i].Email,
				found[i].Email,
			)
		}

		if wanted[i].Type != found[i].Type {
			return fmt.Errorf(
				"Notifications[%d].Type: wanted `%s`; found `%s`",
				i,
				wanted[i].Type,
				found[i].Type,
			)
		}

		if wanted[i].User != found[i].User {
			return fmt.Errorf(
				"Notifications[%d].User: wanted `%s`; found `%s`",
				i,
				wanted[i].User,
				found[i].User,
			)
		}

		wantedClaims, err := parseClaims(key, wanted[i].Token)
		if err != nil {
			return fmt.Errorf(
				"Notifications[%d].Token: parsing wanted token: %w",
				i,
				err,
			)
		}

		foundClaims, err := parseClaims(key, found[i].Token)
		if err != nil {
			return fmt.Errorf(
				"Notifications[%d].Token: parsing found token: %w",
				i,
				err,
			)
		}

		if *wantedClaims != *foundClaims {
			wanted, err := json.Marshal(wantedClaims)
			if err != nil {
				return fmt.Errorf(
					"marshaling wanted[%d]'s token claims: %w",
					i,
					err,
				)
			}
			found, err := json.Marshal(foundClaims)
			if err != nil {
				return fmt.Errorf(
					"marshaling found[%d]'s token claims: %w",
					i,
					err,
				)
			}
			return fmt.Errorf(
				"Notifications[%d].Token: wanted `%s`; found `%s`",
				i,
				wanted,
				found,
			)
		}
	}

	return nil
}

func parseClaims(key ecdsa.PublicKey, tok string) (*types.Claims, error) {
	var claims types.Claims
	if _, err := jwt.ParseWithClaims(
		tok,
		&claims,
		func(*jwt.Token) (interface{}, error) { return &key, nil },
	); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	return &claims, nil
}

type response struct {
	status  int
	body    testsupport.WantedBody
	headers http.Header
}

func (wanted *response) compare(found pz.Response) error {
	if wanted.status == 0 {
		wanted.status = http.StatusOK
	}
	if wanted.status != found.Status {
		data, err := pztest.ReadAll(found.Data)
		if err == nil {
			return fmt.Errorf(
				"response: status: wanted `%d`; found `%d` (body: %s)",
				wanted.status,
				found.Status,
				data,
			)
		}
		return fmt.Errorf(
			"response: status: wanted `%d`; found `%d` (error reading "+
				"response body: %w)",
			wanted.status,
			found.Status,
			err,
		)
	}

	if err := wanted.body(found.Data); err != nil {
		return fmt.Errorf("response: body: %w", err)
	}

	if wanted.headers == nil {
		wanted.headers = http.Header{}
	}
	if found.Headers == nil {
		found.Headers = http.Header{}
	}
	if len(wanted.headers) == len(found.Headers) {
		for h, wanted := range wanted.headers {
			found := found.Headers[h]
			if len(wanted) == len(found) {
				for i := range wanted {
					if wanted[i] != found[i] {
						goto ERROR
					}
				}
				return nil
			}
		ERROR:
			return fmt.Errorf(
				"response: headers: `%s`: wanted `[\"%s\"]`; found `[\"%s\"]`",
				h,
				strings.Join(wanted, "\", \""),
				strings.Join(found, "\", \""),
			)
		}
	}

	return nil
}
