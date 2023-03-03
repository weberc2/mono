package main

import (
	"log"
	"net/http"
	"net/netip"
	"os"

	"github.com/weberc2/mono/mod/nodeinit/pkg/model"
	"github.com/weberc2/mono/mod/nodeinit/pkg/server"
	"tailscale.com/client/tailscale"
)

func main() {
	tailscale.I_Acknowledge_This_API_Is_Unstable = true
	tailnet := os.Getenv("TAILNET")
	if tailnet == "" {
		log.Fatalln("FATAL missing required env var `TAILNET`")
	}

	tailscaleAPIKey := os.Getenv("TAILSCALE_API_KEY")
	if tailscaleAPIKey == "" {
		log.Fatalln("FATAL missing required env var `TAILSCALE_API_KEY`")
	}

	server := server.New(
		tailscale.NewClient(tailnet, tailscale.APIKey(tailscaleAPIKey)),
		model.MemoryNodeStore{
			netip.MustParseAddr("192.168.68.100"): &model.Node{},
			netip.MustParseAddr("192.168.68.58"): &model.Node{
				Hostname: "client",
				Tags:     []string{"tag:server"},
			},
			netip.MustParseAddr("172.19.0.3"): &model.Node{
				Hostname: "docker-client",
				Tags:     []string{"tag:server"},
			},
		},
	)

	addr := ":8080"
	port := os.Getenv("PORT")
	if port != "" {
		addr = ":" + port
	}
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("listening at `%s`: %v", addr, err)
	}
}
