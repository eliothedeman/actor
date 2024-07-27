package actor

import (
	"sync"
)

type mailbox[T any] struct {
	*sync.Cond
	data   []T
	closed bool
}

func newMailBox[T any]() *mailbox[T] {
	m := new(mailbox[T])
	m.Cond = sync.NewCond(&sync.Mutex{})
	return m
}

func (m *mailbox[T]) pushBack(d T) {
	m.data = append(m.data, d)
}

func (m *mailbox[T]) peekFront() T {
	return m.data[0]
}

func (m *mailbox[T]) popFront() T {
	t := m.data[0]
	m.data = m.data[1:]
	return t
}

func (m *mailbox[T]) len() int {
	return len(m.data)
}

func (m *mailbox[T]) close() {
	m.L.Lock()
	defer m.L.Unlock()
	m.closed = true
	m.Broadcast()
}

func (m *mailbox[T]) send(v T) {
	m.L.Lock()
	defer m.L.Unlock()
	m.pushBack(v)
	m.Signal()
}

func (m *mailbox[T]) recieve() T {
	m.L.Lock()
	defer m.L.Unlock()
	for m.len() < 1 {
		m.Wait()
		if m.closed {
			panic("closed")
		}
	}
	return m.popFront()
}
