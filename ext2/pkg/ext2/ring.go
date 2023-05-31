package ext2

// Ring is a temporary, very naive ring buffer implementation until I can
// invest time in a more serious implementation.
type Ring struct {
	buf []Ino
}

func NewRing() Ring { return Ring{buf: nil} }

func (r *Ring) PushBack(value Ino) {
	r.buf = append(r.buf, value)
}

func (r *Ring) PopFront() (Ino, bool) {
	if len(r.buf) < 1 {
		return 0, false
	}

	ret := r.buf[0]
	copy(r.buf, r.buf[1:])
	r.buf = r.buf[:len(r.buf)-1]
	return ret, true
}
