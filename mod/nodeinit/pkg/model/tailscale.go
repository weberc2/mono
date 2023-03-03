package model

import (
	"context"

	"tailscale.com/client/tailscale"
)

type Tailscale interface {
	CreateKey(
		ctx context.Context,
		capabilities tailscale.KeyCapabilities,
	) (string, *tailscale.Key, error)
}
