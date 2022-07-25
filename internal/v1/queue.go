package v1

type node[T any] struct {
	value T
	next  *node[T]
}

type Queue[T any] struct {
	head *node[T]
	tail *node[T]
}

func (q *Queue[T]) PushHead(value T) {
	n := &node[T]{
		value: value,
	}
	if q.head != nil {
		q.head.next = n
	} else {
		q.tail = n
	}
	q.head = n
}

func (q *Queue[T]) CutTail() {
	if q.tail == nil {
		return
	}
	q.tail = q.tail.next
	if q.tail == nil {
		q.head = nil
	}
}

func (q Queue[T]) TailValue() (T, bool) {
	if q.tail != nil {
		return q.tail.value, true
	}
	return *new(T), false
}
