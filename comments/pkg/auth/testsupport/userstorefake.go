package testsupport

import (
	"fmt"

	"github.com/weberc2/mono/comments/pkg/auth/types"
)

type UserStoreFake map[types.UserID]*types.UserEntry

func (usf UserStoreFake) Get(u types.UserID) (*types.UserEntry, error) {
	if entry, found := usf[u]; found {
		return entry, nil
	}
	return nil, types.ErrUserNotFound
}

func (usf UserStoreFake) Insert(entry *types.UserEntry) error {
	usf[entry.User] = entry
	return nil
}

func (usf UserStoreFake) Update(entry *types.UserEntry) error {
	if _, ok := usf[entry.User]; ok {
		usf[entry.User] = entry
		return nil
	}
	return types.ErrUserNotFound
}

func (usf UserStoreFake) List() []*types.UserEntry {
	entries := make([]*types.UserEntry, 0, len(usf))
	for _, entry := range usf {
		entries = append(entries, entry)
	}
	return entries
}

func (usf UserStoreFake) ExpectUsers(wanted []types.Credentials) error {
	if len(usf) != len(wanted) {
		return fmt.Errorf(
			"validating users: length mismatch: wanted `%d` users; found `%d`",
			len(wanted),
			len(usf),
		)
	}

	for _, creds := range wanted {
		entry, ok := usf[creds.User]
		if !ok {
			return fmt.Errorf(
				"validating users: missing expected user: `%s`",
				creds.User,
			)
		}
		if err := creds.CompareUserEntry(entry); err != nil {
			return fmt.Errorf(
				"validating users: mismatch for user `%s`: %w",
				creds.User,
				err,
			)
		}
	}

	return nil
}
