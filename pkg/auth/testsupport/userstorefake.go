package testsupport

import "github.com/weberc2/mono/pkg/auth/types"

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

func (usf UserStoreFake) Upsert(entry *types.UserEntry) error {
	usf[entry.User] = entry
	return nil
}

func (usf UserStoreFake) List() []*types.UserEntry {
	entries := make([]*types.UserEntry, 0, len(usf))
	for _, entry := range usf {
		entries = append(entries, entry)
	}
	return entries
}
