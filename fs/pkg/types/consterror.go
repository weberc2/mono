package types

type ConstError string

func (err ConstError) Error() string { return string(err) }
