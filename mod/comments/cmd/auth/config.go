package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/kelseyhightower/envconfig"
	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/mod/comments/pkg/auth"
	"github.com/weberc2/mono/mod/comments/pkg/auth/types"
	"github.com/weberc2/mono/mod/comments/pkg/pgtokenstore"
	"github.com/weberc2/mono/mod/comments/pkg/pguserstore"
	"gopkg.in/yaml.v2"
)

const (
	envVarPrefix = "AUTH"
	appName      = "auth"
)

type Config struct {
	Addr                    string     `envconfig:"AUTH_ADDR"                      default:"127.0.0.1:8080" yaml:"addr"`
	HostName                string     `envconfig:"AUTH_HOST_NAME"                                          yaml:"hostName"`
	Issuer                  string     `envconfig:"AUTH_ISSUER"                                             yaml:"issuer"`
	Audience                string     `envconfig:"AUTH_AUDIENCE"                                           yaml:"audience"`
	CodeSigningKey          PrivateKey `envconfig:"AUTH_CODE_SIGNING_KEY"                                   yaml:"codeSigningKey"`
	AccessSigningKey        PrivateKey `envconfig:"AUTH_ACCESS_SIGNING_KEY"                                 yaml:"accessSigningKey"`
	RefreshSigningKey       PrivateKey `envconfig:"AUTH_REFRESH_SIGNING_KEY"                                yaml:"refreshSigningKey"`
	ResetSigningKey         PrivateKey `envconfig:"AUTH_RESET_SIGNING_KEY"                                  yaml:"resetSigningKey"`
	NotificationSender      string     `envconfig:"AUTH_NOTIFICATION_SENDER"                                yaml:"notificationSender"`
	DefaultRedirectLocation string     `envconfig:"AUTH_DEFAULT_REDIRECT_LOCATION"                          yaml:"defaultRedirectLocation"`
	RedirectDomain          string     `envconfig:"AUTH_REDIRECT_DOMAIN"                                    yaml:"redirectDomain"`
	BaseURL                 BaseURL    `envconfig:"AUTH_BASE_URL"                                           yaml:"baseURL"`
}

func LoadConfig() (*Config, error) {
	configFile := os.Getenv(envVarPrefix + "_CONFIG_FILE")
	if configFile == "" {
		configFile = filepath.Join(
			os.Getenv("USER"),
			".config",
			appName+".yaml",
		)
	}

	var c Config
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config file: %w", err)
		}

		if err := yaml.UnmarshalStrict(data, &c); err != nil {
			return nil, fmt.Errorf("unmarshaling config file: %w", err)
		}
	}

	if err := envconfig.Process(envVarPrefix, &c); err != nil {
		return nil, fmt.Errorf("parsing environment variables: %w", err)
	}

	return &c, nil
}

func (c *Config) Validate() error {
	if y, e := func() (string, string) {
		if c.Addr == "" {
			return "addr", "ADDR"
		}
		if c.HostName == "" {
			return "hostName", "HOST_NAME"
		}
		if c.Issuer == "" {
			return "issuer", "ISSUER"
		}
		if c.Audience == "" {
			return "audience", "AUDIENCE"
		}
		if c.CodeSigningKey == (PrivateKey{}) {
			return "codeSigningKey", "CODE_SIGNING_KEY"
		}
		if c.AccessSigningKey == (PrivateKey{}) {
			return "accessSigningKey", "ACCESS_SIGNING_KEY"
		}
		if c.RefreshSigningKey == (PrivateKey{}) {
			return "refreshSigningKey", "REFRESH_SIGNING_KEY"
		}
		if c.ResetSigningKey == (PrivateKey{}) {
			return "resetSigningKey", "RESET_SIGNING_KEY"
		}
		if c.NotificationSender == "" {
			return "notificationSender", "NOTIFICATION_SENDER"
		}
		if c.RedirectDomain == "" {
			return "redirectDomain", "REDIRECT_DOMAIN"
		}
		if c.BaseURL == "" {
			return "baseURL", "BASE_URL"
		}
		return "", ""
	}(); y != "" {
		return fmt.Errorf(
			"missing required configuration: %s / %s_%s",
			y,
			envVarPrefix,
			e,
		)
	}
	return nil
}

func (c *Config) Run() error {
	if err := c.Validate(); err != nil {
		return err
	}
	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("creating AWS session: %w", err)
	}

	tokenStore, err := pgtokenstore.OpenEnv()
	if err != nil {
		return fmt.Errorf("opening token store database connection: %w", err)
	}
	if err := tokenStore.EnsureTable(); err != nil {
		return fmt.Errorf("ensuring tokens table exists: %w", err)
	}

	userStore, err := pguserstore.OpenEnv()
	if err != nil {
		return fmt.Errorf("opening user store database connection: %w", err)
	}
	if err := userStore.EnsureTable(); err != nil {
		return fmt.Errorf("ensuring users table exists: %w", err)
	}

	authService := auth.AuthHTTPService{
		AuthService: auth.AuthService{
			Tokens: tokenStore,
			Creds:  auth.CredStore{Users: userStore},
			Codes: types.TokenFactory{
				Issuer:        c.Issuer,
				Audience:      c.Audience,
				TokenValidity: time.Minute,
				SigningKey:    c.CodeSigningKey.Std(),
			},
			ResetTokens: types.ResetTokenFactory{
				Issuer:        c.Issuer,
				Audience:      c.Audience,
				TokenValidity: 1 * time.Hour,
				SigningKey:    c.ResetSigningKey.Std(),
			},
			Notifications: &auth.SESNotificationService{
				Client: ses.New(sess),
				Sender: c.NotificationSender,
				RegistrationSettings: auth.DefaultRegistrationSettings(
					"https://" + strings.TrimRight(c.HostName, "/"),
				),
				ForgotPasswordSettings: auth.DefaultForgotPasswordSettings(
					"https://" + strings.TrimRight(c.HostName, "/"),
				),
			},
			TokenDetails: auth.TokenDetailsFactory{
				AccessTokens: types.TokenFactory{
					Issuer:        c.Issuer,
					Audience:      c.Audience,
					TokenValidity: 15 * time.Minute,
					SigningKey:    c.AccessSigningKey.Std(),
				},
				RefreshTokens: types.TokenFactory{
					Issuer:        c.Issuer,
					Audience:      c.Audience,
					TokenValidity: 7 * 24 * time.Hour,
					SigningKey:    c.RefreshSigningKey.Std(),
				},
				TimeFunc: time.Now,
			},
			TimeFunc: time.Now,
		},
	}

	webServer := auth.WebServer{
		AuthService:             authService.AuthService,
		BaseURL:                 c.BaseURL.Std(),
		RedirectDomain:          c.RedirectDomain,
		DefaultRedirectLocation: c.DefaultRedirectLocation,
	}

	log.Printf(`{"message": "listening on %s"}`, c.Addr)
	if err := http.ListenAndServe(
		c.Addr,
		pz.Register(
			pz.JSONLog(os.Stderr),
			append(
				authService.Routes(),
				webServer.Routes()...,
			)...,
		),
	); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}
	return nil
}

type BaseURL string

func (burl *BaseURL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return burl.Decode(s)
}

func (burl *BaseURL) Decode(value string) error {
	*burl = BaseURL(value)
	if value[len(value)-1] != '/' {
		*burl = *burl + "/"
	}
	return nil
}

func (burl BaseURL) Std() string {
	return string(burl)
}

type PrivateKey ecdsa.PrivateKey

func (pk *PrivateKey) Decode(value string) error {
	data := []byte(value)

	for {
		block, rest := pem.Decode(data)
		if block == nil {
			return fmt.Errorf("input isn't PEM data")
		}
		// Ideally we would just match on PRIVATE KEY, but Terraform's
		// tls_private_key[0] module uses "EC PRIVATE KEY" 🤦
		//
		// [0]:
		// https://registry.terraform.io/providers/hashicorp/tls/latest/docs/resources/private_key#attributes-reference
		if !strings.Contains(block.Type, "PRIVATE KEY") {
			if len(rest) > 0 {
				data = rest
				continue
			}
			return fmt.Errorf("PEM data is missing a 'PRIVATE KEY' block")
		}
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("parsing ecdsa private key: %w", err)
		}
		*pk = PrivateKey(*key)
		return nil
	}
}

func (pk *PrivateKey) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return fmt.Errorf("yaml-unmarshaling *PrivateKey: %w", err)
	}

	if err := pk.Decode(s); err != nil {
		return fmt.Errorf("yaml-unmarshaling *PrivateKey: %w", err)
	}

	return nil
}

func (pk *PrivateKey) Std() *ecdsa.PrivateKey {
	return (*ecdsa.PrivateKey)(pk)
}
