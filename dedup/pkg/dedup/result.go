package dedup

type Result[T any] struct {
	OK  T
	Err error
}
