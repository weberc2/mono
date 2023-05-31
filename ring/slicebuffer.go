package ring

import (
	"errors"
	"fmt"
)

type SliceBuffer[T any] struct {
	// entries holds the items. there is always one more entry than items
	// because there is a 'spacer' slot preceding `start` which allows us to
	// disambiguate between a full buffer (`tail == spacer`) and an empty
	// buffer (`tail == start`).
	entries []T

	// start always points to the beginning of the list. the slot preceding
	// start is always considered invalid.
	start int

	// tail always points to the end of the list. if the list is empty, then
	// start == tail. if tail points to the slot preceding start, then the list
	// is full.
	tail int
}

func NewSliceBuffer[T any](capacity int) (*SliceBuffer[T], error) {
	if capacity < 1 {
		return nil, fmt.Errorf(
			"capacity `%d`: %w",
			capacity,
			NonPositiveCapacityErr,
		)
	}
	return &SliceBuffer[T]{entries: make([]T, capacity+1)}, nil
}

func (buf *SliceBuffer[T]) Cap() int { return len(buf.entries) - 1 }

func (buf *SliceBuffer[T]) Full() bool {
	return (buf.start > 0 && buf.tail == 0) ||
		(buf.start == 0 && buf.tail == len(buf.entries)-2)
}

func (buf *SliceBuffer[T]) Len() int {
	// S
	// [_, #]
	// T
	//
	// S
	// [X, #]
	//    T
	if buf.tail >= buf.start {
		return buf.tail - buf.start
	}

	//    S
	// [#, X]
	// T
	//
	//       S
	// [X, #, X]
	//    T
	//
	//
	//             S
	// [X, X, _, #, X] len(entries)=5, S=4, T=2, Len()=5-(4-2)=5-2=3
	//       T
	return len(buf.entries) - (buf.start - buf.tail)
}

func (buf *SliceBuffer[T]) next(i int) int {
	return (i + 1) % len(buf.entries)
}

func (buf *SliceBuffer[T]) prev(i int) int {
	// we are adding `len(buf.entries)` to the numerator so that we get Python's
	// wraparound behavior (e.g., `-1 % n -> n-1`) instead of Go's
	// `-1 % n -> -1`. Of course, this only works so long as
	// `i < len(buf.entries)`, but that should always be the case.
	return (len(buf.entries) + i - 1) % len(buf.entries)
}

// what is the difference between (cap=1, len=0) [ST] and (cap=1, len=1) [ST]
func (buf *SliceBuffer[T]) Push(item T) (T, bool) {
	// S            S
	// [X, #] => [#, X]
	//    T      T
	//
	//    S      S
	// [#, X] => [X, #]
	// T            T
	//
	//       S      S
	// [X, #, X] => [X, X, #]
	//    T               T
	if spacer := buf.prev(buf.start); buf.tail == spacer {
		evicted := buf.entries[buf.start]
		buf.entries[spacer] = item
		buf.tail = buf.start            // advance tail by 1
		buf.start = buf.next(buf.start) // advance start by 1
		return evicted, true
	}

	//
	// S
	// [X, _, #]
	//    T
	//
	//    S            S
	// [#, X, _] => [#, X, X]
	//       T      T
	//             S                  S
	// [X, X, _, #, X] => [X, X, X, #, X]
	//       T                     T
	buf.entries[buf.tail] = item
	buf.tail = buf.next(buf.tail)
	var zero T
	return zero, false
}

func (buf *SliceBuffer[T]) PopFront() (T, bool) {
	// empty
	if buf.tail == buf.start {
		var zero T
		return zero, false
	}

	popped := buf.entries[buf.start]
	buf.start = buf.next(buf.start)
	return popped, true
}

func (buf *SliceBuffer[T]) Items() []T {
	items := make([]T, buf.Len())
	cap := buf.Cap()
	for i := range items {
		items[i] = buf.entries[buf.start+i%cap]
	}
	return items
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

var NonPositiveCapacityErr = errors.New("capacity must be positive")
