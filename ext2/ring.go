package main

type Ring struct {
	Next, Prev *Ring
	Value      uint64
}

func NewRing(value uint64) *Ring {
	r := new(Ring)
	r.Next = r
	r.Prev = r
	return r
}

func (r *Ring) Push(value uint64) *Ring {
	if r == nil {
		return NewRing(value)
	}

	newRing := &Ring{
		Value: value,
		Prev:  r,
		Next:  r.Next,
	}

	r.Next.Prev = newRing
	r.Next = newRing
	return r
}
