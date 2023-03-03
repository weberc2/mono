package model

import (
	"net/netip"
)

type Node struct {
	Hostname string   `json:"hostname"`
	Tags     []string `json:"tags"`
}

type NodeStore interface {
	GetNode(ip netip.Addr) (*Node, error)
}
