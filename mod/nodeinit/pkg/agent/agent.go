package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/weberc2/mono/mod/nodeinit/pkg/client"
)

const dataDir = "/var/lib/nodeinit/agent/"
const hasRunBeforeFile = dataDir + "hasrunbefore"

type Agent struct {
	Client client.Client
}

func New() *Agent { return &Agent{*client.New()} }

func (agent *Agent) Run(ctx context.Context) error {
	log.Print("DEBUG ensuring program is running as root user")
	if err := agent.EnsureRoot(); err != nil {
		return fmt.Errorf("running agent: %w", err)
	}

	log.Print("DEBUG ensuring the agent hasn't already run")
	firstRun, err := agent.IsFirstRun()
	if err != nil {
		return fmt.Errorf("running agent: ensuring first run: %w", err)
	}
	if !firstRun {
		log.Printf("INFO agent has already run; aborting.")
		return nil
	}

	log.Print("DEBUG ensuring tailscale is installed")
	if err := agent.EnsureTailscale(ctx); err != nil {
		return fmt.Errorf("running agent: %w", err)
	}

	log.Print("DEBUG fetching user-data")
	userData, err := agent.Client.GetUserData(ctx)
	if err != nil {
		return fmt.Errorf("running agent: %w", err)
	}

	log.Printf("DEBUG setting hostname to `%s`", userData.Hostname)
	if err := Sethostname([]byte(userData.Hostname)); err != nil {
		return fmt.Errorf(
			"running agent: setting hostname to `%s`: %w",
			userData.Hostname,
			err,
		)
	}

	log.Print("DEBUG running `tailscale up ...`")
	if err := agent.TailscaleUp(userData.TailscaleAuthKey); err != nil {
		return fmt.Errorf("running agent: %w", err)
	}

	log.Printf("DEBUG touching `%s`", hasRunBeforeFile)
	if err := agent.TouchHasRunBefore(); err != nil {
		return fmt.Errorf("running agent: %w", err)
	}
	return nil
}

func (agent *Agent) EnsureRoot() error {
	if os.Geteuid() != 0 {
		return errors.New("not running as root user")
	}
	return nil
}

func (agent *Agent) TailscaleUp(key string) error {
	if err := runCmd(
		"[tailscale]",
		"tailscale",
		"up",
		"--ssh",
		"--auth-key",
		key,
	); err != nil {
		return fmt.Errorf("running `tailscale up ...`: %v", err)
	}

	return nil
}

func (agent *Agent) IsFirstRun() (bool, error) {
	_, err := os.Stat(hasRunBeforeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (agent *Agent) TouchHasRunBefore() error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf(
			"touching `%s`: creating parent directories: %w",
			hasRunBeforeFile,
			err,
		)
	}
	f, err := os.Create(hasRunBeforeFile)
	if err != nil {
		return fmt.Errorf("touching `%s`: %w", hasRunBeforeFile, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf(
			"touching `%s`: closing file: %w",
			hasRunBeforeFile,
			err,
		)
	}
	return nil
}

func (agent *Agent) EnsureTailscale(ctx context.Context) error {
	if err := runCmd(
		"[ensuring tailscale]",
		"/bin/sh",
		"-c",
		"command -v tailscale",
	); err == nil {
		// if err is nil, then tailscale is already installed
		return nil
	}

	// otherwise install it
	if err := addAptRepo(
		ctx,
		downloadParams{
			url:  "https://pkgs.tailscale.com/stable/ubuntu/jammy.noarmor.gpg",
			file: "/usr/share/keyrings/tailscale-archive-keyring.gpg",
		},
		downloadParams{
			url:  "https://pkgs.tailscale.com/stable/ubuntu/jammy.tailscale-keyring.list",
			file: "/etc/apt/sources.list.d/tailscale.list",
		},
	); err != nil {
		return fmt.Errorf("ensuring tailscale installed: %w", err)
	}
	if err := runCmd(
		"[install tailscale]",
		"apt-get",
		"install",
		"-y",
		"tailscale",
	); err != nil {
		return fmt.Errorf(
			"ensuring tailscale: tailscale not found so installing: %w",
			err,
		)
	}
	return nil
}
