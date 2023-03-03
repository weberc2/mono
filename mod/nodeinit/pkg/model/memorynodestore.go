package model

import (
	"net/netip"

	"github.com/weberc2/mono/mod/nodeinit/pkg/protocol"
)

// MemoryNodeStore is an in-memory representation of `NodeStore`
type MemoryNodeStore map[netip.Addr]*Node

// GetNode implements `NodeStore.GetNode()`
func (store MemoryNodeStore) GetNode(ip netip.Addr) (*Node, error) {
	node, found := store[ip]
	if !found {
		return nil, &protocol.NodeNotFoundErr{IP: ip}
	}

	return node, nil
}
