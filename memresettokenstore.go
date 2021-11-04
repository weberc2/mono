package main

type MemResetTokenStore struct {
	Tokens []*ResetToken
}

func (mrts *MemResetTokenStore) Create(rt *ResetToken) error {
	for i, tok := range mrts.Tokens {
		if tok.User == rt.User {
			mrts.Tokens[i] = rt
			return nil
		}
	}
	mrts.Tokens = append(mrts.Tokens, rt)
	return nil
}

func (mrts *MemResetTokenStore) Get(user UserID) (*ResetToken, error) {
	for _, tok := range mrts.Tokens {
		if tok.User == user {
			return tok, nil
		}
	}
	return nil, ErrResetTokenNotFound
}
