package main

type MemUserStore struct {
	Entries []*UserEntry
}

func (mus *MemUserStore) Create(entry *UserEntry) error {
	for _, e := range mus.Entries {
		if e.User == entry.User {
			return ErrUserExists
		}
	}
	mus.Entries = append(mus.Entries, entry)
	return nil
}

func (mus *MemUserStore) Update(entry *UserEntry) error {
	for i, e := range mus.Entries {
		if e.User == entry.User {
			mus.Entries[i] = entry
			return nil
		}
	}
	return ErrUserNotFound
}

func (mus *MemUserStore) Get(user UserID) (*UserEntry, error) {
	for _, e := range mus.Entries {
		if e.User == user {
			return e, nil
		}
	}
	return nil, ErrUserNotFound
}
