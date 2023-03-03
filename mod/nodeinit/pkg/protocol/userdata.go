package protocol

// UserData represents the user-data configuration for the nodeinit system. It
// contains secrets and should be treated as sensitive data.
type UserData struct {
	// Hostname is the hostname for the node.
	Hostname string `json:"hostname"`

	// TailscaleAuthKey is the tailscale auth key for the node. This is a
	// secret and should be treated as sensitive data.
	TailscaleAuthKey string `json:"tailscaleAuthKey"`
}
