package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/weberc2/mono/mod/nodeinit/pkg/protocol"
)

// ServerAddr holds the server address for the nodeinit server.
const ServerAddr = "192.168.68.100"

// GetUserData fetches the user-data with the default client.
func GetUserData(ctx context.Context) (*protocol.UserData, error) {
	return (&Client{}).GetUserData(ctx)
}

// Client is a client for the nodeinit server.
type Client struct {
	// HTTP holds the HTTP client to use to communicate with the nodeinit
	// server.
	HTTP       http.Client
	ServerAddr string
}

func New() *Client {
	var httpClient http.Client
	httpClient.Timeout = 5 * time.Second
	return &Client{HTTP: httpClient, ServerAddr: ServerAddr}
}

func (client *Client) SetServerAddr(addr string) *Client {
	client.ServerAddr = addr
	return client
}

// GetUserData fetches the user-data from the nodeinit server.
func (client *Client) GetUserData(
	ctx context.Context,
) (*protocol.UserData, error) {
	req, err := http.NewRequest("GET", "http://"+client.ServerAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching user-data: creating request: %v", err)
	}
	rsp, err := client.HTTP.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf(
			"fetching user-data: issuing HTTP GET request to `%s`: %v",
			client.ServerAddr,
			err,
		)
	}

	defer rsp.Body.Close()
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"fetching user-data: reading response body: %v",
			err,
		)
	}

	if rsp.StatusCode != http.StatusOK {
		var protocolErr protocol.Error
		if err := json.Unmarshal(data, &protocolErr); err == nil {
			return nil, &protocolErr
		}
		return nil, fmt.Errorf(
			"fetching user-data: received unexpected status code `%d`",
			rsp.StatusCode,
		)
	}

	var userdata protocol.UserData
	if err := json.Unmarshal(data, &userdata); err != nil {
		data, err := json.Marshal(struct {
			Level          string `json:"level"`
			Data           string `json:"data"`
			UnmarshalError string `json:"unmarshalError"`
			Context        string `json:"context"`
		}{
			Level:          "ERROR",
			Data:           string(data),
			UnmarshalError: err.Error(),
			Context:        "fetching user-data: unmarshaling JSON",
		})
		if err != nil {
			panic(fmt.Sprintf("marshaling error response as json: %v", err))
		}
		log.Printf("%s", data)
		return nil, fmt.Errorf(
			"fetching user-data: unmarshaling JSON: %v",
			err,
		)
	}

	return &userdata, nil
}
