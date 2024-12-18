package dedup

type Option[T any] struct {
	Exists bool
	Some   T
}

func Some[T any](some T) (opt Option[T]) {
	opt.Exists = true
	opt.Some = some
	return
}

func OptionMap[F, T any](in Option[F], mapper func(*F) T) (out Option[T]) {
	if in.Exists {
		out = Some(mapper(&in.Some))
	}
	return
}
