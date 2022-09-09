package iters

type Collection[T comparable] []T

// CallbackIter is an iterator function with callback. This approach is the
// most straight forward.
func (c Collection[T]) CallbackIter(fn func(v T) bool) {
	for _, val := range c {
		if !fn(val) {
			break
		}
	}
}

// ClosureIter is an iterator function that returns a closure. The returned
// iterator closure is a generator function that you call repeatedly.
func (c Collection[T]) ClosureIter() (func() (val T, next bool), bool) {
	var idx int
	return func() (T, bool) {
		prev := idx
		idx++
		return c[prev], idx < len(c)
	}, idx < len(c)
}

// ChannelIter simply uses channels to create an iterator. The one thing to
// keep in mind is that you always have to make sure to drain the channel
// if you exit early.
func (c Collection[T]) ChannelIter(bufSize int) <-chan T {
	ch := make(chan T, bufSize)
	go func() {
		for _, val := range c {
			ch <- val
		}
		close(ch)
	}()
	return ch
}

// StatefulIter is a bit like the ClosureIter, but it is responsible for
// storing and maintaining the state, as opposed to the ClosureIterators
// current state, which is implicit. It has a constructor and two methods,
// Value() and Next().
type StatefulIter[T comparable] struct {
	current int
	data    []T
}

// NewStatefulIter creates and returns a new stateful iterator instance.
func NewStatefulIter[T comparable](data Collection[T]) *StatefulIter[T] {
	return &StatefulIter[T]{
		current: 0,
		data:    data,
	}
}

// Value returns the value of the current state of the iterator.
func (s *StatefulIter[T]) Value() T {
	return s.data[s.current]
}

// Next advances the iterator and reports if there is still more data to
// be processed on future calls.
func (s *StatefulIter[T]) Next() bool {
	s.current++
	return s.current < len(s.data)
}
