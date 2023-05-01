package fs

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Integer interface {
	Signed | Unsigned
}

func Min[T Integer](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T Integer](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func DivCiel[T Integer](a, b T) T {
	if a%b > 0 {
		return a/b + 1
	}
	return a / b
}
