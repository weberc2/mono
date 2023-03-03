package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
)

const (
	typeNodeNotFoundErr = "NodeNotFoundErr"
)

type Error struct {
	Type    string          `json:"type"`
	Message string          `json:"message"`
	Err     json.RawMessage `json:"error"`
}

func (e *Error) Error() string { return e.Message }

func (e *Error) JSON() []byte {
	data, err := json.Marshal(e)
	if err != nil {
		panic(fmt.Sprintf("marshaling `Error` as JSON: %v", err))
	}
	return data
}

func (e *Error) NodeNotFoundErr() (*NodeNotFoundErr, error) {
	if e.Type != typeNodeNotFoundErr {
		return nil, ErrNotNodeNotFoundErr
	}
	var nodeNotFoundErr NodeNotFoundErr
	if err := json.Unmarshal(e.Err, &nodeNotFoundErr.IP); err != nil {
		return nil, fmt.Errorf(
			"unmarshaling `NodeNotFoundErr` from JSON: %s: %v",
			"parsing payload's `error` field as `NodeNotFoundErr`",
			err,
		)
	}
	return &nodeNotFoundErr, nil
}

var ErrNotNodeNotFoundErr = errors.New("not a NodeNotFoundErr")

type NodeNotFoundErr struct {
	IP netip.Addr
}

func (err *NodeNotFoundErr) Error() string {
	return fmt.Sprintf("node `%s` not found", err.IP)
}

func (err *NodeNotFoundErr) ProtocolErr() *Error {
	ipdata, e := json.Marshal(err.IP)
	if e != nil {
		panic(fmt.Sprintf(
			"marshaling ip address `%s` to JSON: %v",
			err.IP,
			e,
		))
	}
	return &Error{
		Type:    "NodeNotFoundErr",
		Message: err.Error(),
		Err:     ipdata,
	}
}
