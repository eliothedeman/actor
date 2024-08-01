package actor

import (
	"runtime"
	"sync"

	"github.com/gammazero/deque"
)

type mailbox[T any] struct {
	sync.Mutex
	data   deque.Deque[T]
	closed bool
}

func newMailBox[T any]() *mailbox[T] {
	m := new(mailbox[T])
	return m
}

func (m *mailbox[T]) len() int {
	return m.data.Len()
}

func (m *mailbox[T]) close() {
	m.Lock()
	defer m.Unlock()
	m.closed = true
}

func (m *mailbox[T]) send(v T) {
	m.Lock()
	defer m.Unlock()
	m.data.PushBack(v)
}

func (m *mailbox[T]) recieve() T {
	for {
		m.Lock()
		if m.len() > 0 {
			x := m.data.PopFront()
			m.Unlock()

			return x
		}
		m.Unlock()
		runtime.Gosched()
	}
}
