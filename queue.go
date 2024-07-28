package actor

import (
	"sync"

	"github.com/gammazero/deque"
)

type mailbox[T any] struct {
	sync.Cond
	data   deque.Deque[T]
	closed bool
}

func newMailBox[T any]() *mailbox[T] {
	m := new(mailbox[T])
	m.Cond.L = &sync.Mutex{}
	return m
}

func (m *mailbox[T]) len() int {
	return m.data.Len()
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
	m.data.PushBack(v)
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
	return m.data.PopFront()
}
