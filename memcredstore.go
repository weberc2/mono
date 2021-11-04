package main

type MemCredStore struct {
	Credentials []*Credentials
}

func (mcs *MemCredStore) Validate(creds *Credentials) error {
	for _, cs := range mcs.Credentials {
		if *cs == *creds {
			return nil
		}
	}
	return ErrCredentials
}

func (mcs *MemCredStore) Create(creds *Credentials) error {
	for _, cs := range mcs.Credentials {
		if cs.User == creds.User {
			return ErrUserExists
		}
	}
	mcs.Credentials = append(mcs.Credentials, creds)
	return nil
}

func (mcs *MemCredStore) Update(creds *Credentials) error {
	for i, cs := range mcs.Credentials {
		if cs.User == creds.User {
			mcs.Credentials[i] = creds
			return nil
		}
	}
	return ErrUserNotFound
}
