package model

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/weberc2/mono/mod/nodeinit/pkg/protocol"
	"tailscale.com/client/tailscale"
)

// Model is a protocol-agnostic model of a user-data server for the nodeinit
// system.
type Model struct {
	// Tailscale is a tailscale client.
	Tailscale Tailscale

	// NodeStore is the backend storage for node configuration.
	NodeStore NodeStore
}

// GetUserData fetches the userdata for a given node. If the node doesn't
// exist, `NodeNotFoundErr` is returned.
func (model *Model) GetUserData(
	ctx context.Context,
	ip netip.Addr,
) (*protocol.UserData, error) {
	node, err := model.NodeStore.GetNode(ip)
	if err != nil {
		return nil, fmt.Errorf("fetching user data: %w", err)
	}

	key, _, err := model.Tailscale.CreateKey(
		ctx,
		tailscale.KeyCapabilities{
			Devices: tailscale.KeyDeviceCapabilities{
				Create: tailscale.KeyDeviceCreateCapabilities{
					Ephemeral:     false,
					Reusable:      false,
					Preauthorized: true,
					Tags:          node.Tags,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"fetching user data: creating tailscale auth key: %w",
			err,
		)
	}

	return &protocol.UserData{
		Hostname:         node.Hostname,
		TailscaleAuthKey: key,
	}, nil
}
