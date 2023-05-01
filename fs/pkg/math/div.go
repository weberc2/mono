package math

func DivRoundUp[T Integer](a, b T) T {
	if a%b == 0 {
		return a / b
	}
	return a/b + 1
}
