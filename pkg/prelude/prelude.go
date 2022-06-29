package prelude

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return t
}
