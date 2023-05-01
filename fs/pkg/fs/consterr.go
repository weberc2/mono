package fs

type constErr string

func (err constErr) Error() string { return string(err) }
