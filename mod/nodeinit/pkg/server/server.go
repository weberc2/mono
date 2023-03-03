package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"strings"

	"github.com/weberc2/mono/mod/nodeinit/pkg/model"
	"github.com/weberc2/mono/mod/nodeinit/pkg/protocol"
)

type Tailscale = model.Tailscale
type NodeStore = model.NodeStore

type Server struct {
	Model model.Model
}

func New(tailscale Tailscale, nodeStore NodeStore) *Server {
	return &Server{
		Model: model.Model{Tailscale: tailscale, NodeStore: nodeStore},
	}
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	addr := r.RemoteAddr
	colon := strings.IndexByte(addr, ':')
	if colon >= 0 {
		addr = addr[:colon]
	}
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		log.Printf(
			"ERROR failed to parse remote address `%s` as IP; "+
				"this shouldn't happen: %v",
			addr,
			err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, http.StatusText(http.StatusInternalServerError))
		return
	}

	userData, err := server.Model.GetUserData(context.Background(), ip)
	if err != nil {
		var targetErr *protocol.NodeNotFoundErr
		if errors.As(err, &targetErr) {
			log.Printf(
				"INFO received request for unknown ip address `%s`: %v",
				ip,
				err,
			)

			type httpError struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			}

			data, err := json.Marshal(httpError{
				Type:    "NodeNotFoundErr",
				Message: targetErr.Error(),
			})
			if err != nil {
				log.Fatalf("encountered error marshaling `httpError`: %v", err)
			}

			w.WriteHeader(http.StatusNotFound)
			w.Write(data)
			return
		}

		log.Printf("ERROR %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, http.StatusText(http.StatusInternalServerError))
		return
	}

	data, err := json.Marshal(userData)
	if err != nil {
		log.Printf(
			"ERROR failed to marshal user data for node `%s`; "+
				"this shouldn't happen: %v",
			ip,
			err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, http.StatusText(http.StatusInternalServerError))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
