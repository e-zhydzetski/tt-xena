package v2

type Node[T any] struct {
	Value T
	next  *Node[T]
}

type RingBuffer[T any] struct {
	cur *Node[T]
}

func (r *RingBuffer[T]) GetNext() *Node[T] {
	if r.cur == nil {
		return nil
	}
	return r.cur.next
}

func (r *RingBuffer[T]) InsertNext(value T) {
	n := &Node[T]{
		Value: value,
	}
	if r.cur == nil {
		n.next = n
		r.cur = n
		return
	}
	n.next = r.cur.next
	r.cur.next = n
}

func (r *RingBuffer[T]) Next() {
	if r.cur == nil {
		return
	}
	r.cur = r.cur.next
}
