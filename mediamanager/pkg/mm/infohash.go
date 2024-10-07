package mm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type InfoHash struct {
	lowercase string
}

func NewInfoHash(s string) (infoHash InfoHash) {
	infoHash.lowercase = strings.ToLower(s)
	return
}

func (infoHash InfoHash) String() string { return infoHash.lowercase }

func (infoHash *InfoHash) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(infoHash.lowercase)
	if err != nil {
		err = fmt.Errorf("marshaling info hash: %w", err)
	}
	return data, err
}

func (infoHash *InfoHash) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshaling info hash: %w", err)
	}
	*infoHash = NewInfoHash(s)
	return nil
}

func (infoHash InfoHash) Value() driver.Value {
	return infoHash.lowercase
}

func (infoHash *InfoHash) Scan(value interface{}) error {
	if s, ok := value.(string); ok {
		*infoHash = NewInfoHash(s)
		return nil
	}
	return fmt.Errorf(
		"invalid sql type for info hash: "+
			"wanted `string`; found `%[1]T` (%[1]v)",
		value,
	)
}
