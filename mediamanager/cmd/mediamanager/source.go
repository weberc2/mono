package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/cenkalti/rain/rainrpc"
)

type Source struct {
	Type    SourceType `json:"type"`
	Magnet  Magnet     `json:"magnet,omitempty"`
	Torrent Torrent    `json:"torrent,omitempty"`
}

var _ json.Marshaler = (*Source)(nil)

// MarshalJSON implements the `json.Marshaler` interface. This exists because
// `omitempty` doesn't actually work for struct fields (specifically the
// `MetaInfo` field), so we have to do it manually, and we can't just implement
// it for the `MetaInfo` type because `json.Marshaler` can't decide whether or
// not it should be omitted, only its parent struct can decide that).
func (source *Source) MarshalJSON() (data []byte, err error) {
	var v interface{}
	switch source.Type {
	case SourceTypeMagnet:
		v = &struct {
			Type   SourceType `json:"type"`
			Magnet Magnet     `json:"magnet"`
		}{
			Type:   SourceTypeMagnet,
			Magnet: source.Magnet,
		}
	case SourceTypeTorrent:
		v = &struct {
			Type    SourceType `json:"type"`
			Torrent Torrent    `json:"torrent"`
		}{
			Type:    SourceTypeTorrent,
			Torrent: source.Torrent,
		}
	default:
		err = fmt.Errorf(
			"marshaling download spec: invalid source type: %s",
			source.Type,
		)
		return
	}

	data, err = json.Marshal(v)
	if err != nil {
		err = fmt.Errorf("marshaling download spec: %w", err)
	}
	return
}

func (source *Source) InfoHash() metainfo.Hash {
	switch source.Type {
	case SourceTypeMagnet:
		return source.Magnet.InfoHash
	case SourceTypeTorrent:
		return (*metainfo.MetaInfo)(&source.Torrent).HashInfoBytes()
	default:
		panic(fmt.Sprintf("invalid source type: %s", source.Type))
	}
}

func (source *Source) addToRain(
	rain *rainrpc.Client,
	opts *rainrpc.AddTorrentOptions,
) error {
	switch source.Type {
	case SourceTypeMagnet:
		return source.Magnet.addToRain(rain, opts)
	case SourceTypeTorrent:
		return source.Torrent.addToRain(rain, opts)
	default:
		return nil
	}
}

type SourceType string

const (
	SourceTypeMetaInfo SourceType = "METAINFO"
	SourceTypeTorrent  SourceType = "TORRENT"
	SourceTypeMagnet   SourceType = "MAGNET"
)

func (s *SourceType) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*string)(s)); err != nil {
		return fmt.Errorf("unmarshaling source type: %w", err)
	}
	switch *s {
	case SourceTypeMetaInfo, SourceTypeTorrent, SourceTypeMagnet:
		return nil
	}
	return fmt.Errorf("unmarshaling source type: invalid source type: %s", *s)
}

type Magnet metainfo.Magnet

func (m *Magnet) String() string {
	return (*metainfo.Magnet)(m).String()
}

func (m *Magnet) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *Magnet) UnmarshalJSON(data []byte) (err error) {
	var src string
	if err = json.Unmarshal(data, &src); err != nil {
		err = fmt.Errorf("unmarshaling magnet: %w", err)
		return
	}
	if *(*metainfo.Magnet)(m), err = metainfo.ParseMagnetUri(src); err != nil {
		err = fmt.Errorf("unmarshaling magnet: %w", err)
	}
	return
}

func (m *Magnet) addToRain(
	rain *rainrpc.Client,
	opts *rainrpc.AddTorrentOptions,
) error {
	if _, err := rain.AddURI(m.String(), opts); err != nil {
		return fmt.Errorf("adding magnet to rain client: %w", err)
	}
	return nil
}

type Torrent metainfo.MetaInfo

func (t *Torrent) MarshalBencode() ([]byte, error) {
	data, err := bencode.Marshal(t)
	if err != nil {
		err = fmt.Errorf("bencoding torrent: %w", err)
	}
	return data, err
}

func (t *Torrent) MarshalJSON() ([]byte, error) {
	data, err := t.MarshalBencode()
	if err != nil {
		return nil, fmt.Errorf("marshaling torrent to json: %w", err)
	}
	if data, err = json.Marshal(string(data)); err != nil {
		err = fmt.Errorf("marshaling torrent to json: %w", err)
	}
	return data, err
}

func (t *Torrent) UnmarshalJSON(data []byte) (err error) {
	var src string
	var info *metainfo.MetaInfo
	if err = json.Unmarshal(data, &src); err != nil {
		err = fmt.Errorf("unmarshaling torrent from json: %w", err)
	} else if info, err = metainfo.Load(strings.NewReader(src)); err != nil {
		err = fmt.Errorf("unmarshaling torrent from json: %w", err)
	} else {
		*t = *(*Torrent)(info)
	}
	return
}

func (t *Torrent) addToRain(
	rain *rainrpc.Client,
	opts *rainrpc.AddTorrentOptions,
) error {
	data, err := t.MarshalBencode()
	if err != nil {
		return fmt.Errorf("adding torrent to rain client: %w", err)
	}
	if _, err := rain.AddTorrent(bytes.NewReader(data), opts); err != nil {
		return fmt.Errorf("adding torrent to rain client: %w", err)
	}
	return nil
}
