package queue

type mqueue[T any] struct {
	data []T
}

func (m *mqueue[T]) pushBack(d T) {
	m.data = append(m.data, d)
}

func (m *mqueue[T]) peekFront() T {
	return m.data[0]
}

func (m *mqueue[T]) popFront() T {
	t := m.data[0]
	m.data = m.data[1:]
	return t
}

func (m *mqueue[T]) len() int {
	return len(m.data)
}

type MQueue[T any] struct {
	mqueue[T]
	in     chan T
	out    chan T
	filter chan func(T) bool
}

func New[T any]() *MQueue[T] {
	m := &MQueue[T]{
		mqueue: mqueue[T]{
			data: []T{},
		},
		in:     make(chan T),
		out:    make(chan T),
		filter: make(chan func(T) bool),
	}
	go m.run()
	return m
}

func (m *MQueue[T]) Filter(f func(T) bool) {
	m.filter <- f
}

func (m *MQueue[t]) run() {
	for {
		if m.len() < 1 {
			select {
			case d := <-m.in:
				m.pushBack(d)
			case <-m.filter:
			}
			continue
		}
		next := m.peekFront()
		select {
		case t := <-m.in:
			m.pushBack(t)
		case m.out <- next:
			m.popFront()
		case f := <-m.filter:
			l := m.len()
			for i := 0; i < l; i++ {
				d := m.popFront()
				if !f(d) {
					m.pushBack(d)
				}
			}
		}
	}
}

func (m *MQueue[T]) Next() T {
	return <-m.out
}
func (m *MQueue[T]) Out() <-chan T {
	return m.out
}

func (m *MQueue[T]) Send(t T) {
	m.in <- t
}
