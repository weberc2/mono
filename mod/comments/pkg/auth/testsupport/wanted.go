package testsupport

import (
	"fmt"

	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
)

func WantedString(wanted string) WantedBody {
	return WantedData(func(data []byte) error {
		if found := string(data); wanted != found {
			return fmt.Errorf("wanted `%s`; found `%s`", wanted, found)
		}
		return nil
	})
}

func WantedData(f func(data []byte) error) WantedBody {
	return func(s pz.Serializer) error {
		data, err := pztest.ReadAll(s)
		if err != nil {
			return fmt.Errorf("reading serializer: %w", err)
		}
		return f(data)
	}
}

type WantedBody func(pz.Serializer) error
