package testsupport

import "github.com/weberc2/auth/pkg/types"

type UserStoreFake map[types.UserID]*types.UserEntry

func (usf UserStoreFake) Get(u types.UserID) (*types.UserEntry, error) {
	if entry, found := usf[u]; found {
		return entry, nil
	}
	return nil, types.ErrUserNotFound
}

func (usf UserStoreFake) Create(entry *types.UserEntry) error {
	usf[entry.User] = entry
	return nil
}

func (usf UserStoreFake) Upsert(entry *types.UserEntry) error {
	usf[entry.User] = entry
	return nil
}
