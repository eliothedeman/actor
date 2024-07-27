package alist

import (
	"errors"
)

var (
	ErrIndexOutOfRange = errors.New("index out of range")
)

const (
	nodeLength = 8
)

type ArrayList[T any] struct {
	head      *node[T]
	tail      *node[T]
	length    int
	deadHeads int
}

func Make[T any](size int) *ArrayList[T] {
	head := new(node[T])
	a := &ArrayList[T]{
		head:   head,
		tail:   head,
		length: 0,
	}
	a.Resize(size)
	return a
}

func (a *ArrayList[T]) Index(i int) T {
	i = a.trueIndex(i)
	if i >= a.Len() {
		panic(ErrIndexOutOfRange)
	}
	n, x := a.head.forward(i)
	return n.data[x]
}

func (a *ArrayList[T]) trueIndex(i int) int {
	return a.deadHeads + i
}
func (a *ArrayList[T]) front() int {
	return a.trueIndex(0)
}

func (a *ArrayList[T]) end() int {
	return a.trueIndex(a.Len() - 1)
}

func (a *ArrayList[T]) PopFront() T {
	v := a.Index(0)
	a.deadHeads++
	a.length--
	if a.deadHeads == nodeLength {
		a.head = a.head.next
		a.deadHeads = 0
	}
	return v
}

func (a *ArrayList[T]) Append(val T) {
	mod := a.length % nodeLength
	t := a.tail
	have := nodesFor(a.length)
	want := nodesFor(a.length + 1)
	if have != want {
		t.next = new(node[T])
		t = t.next
		a.tail = t
	}
	t.data[mod] = val
	a.length++
}

func (a *ArrayList[T]) Insert(at int, val T) {
	at = a.trueIndex(at)
	if a.length <= at {
		panic(ErrIndexOutOfRange)
	}
	n, x := a.head.forward(at)
	n.data[x] = val
}

func nodesFor(size int) int {
	return (size / nodeLength) + 1
}

func (a *ArrayList[T]) upsize(size int) {
	have := nodesFor(a.length)
	want := nodesFor(size)
	a.length = size
	if have == want {
		return
	}
	t := a.tail
	for i := 0; i < want-have; i++ {
		t.next = new(node[T])
		t = t.next
	}
	a.tail = t
}

func (a *ArrayList[T]) downsize(size int) {
	n, _ := a.head.forward(size)
	n.next = nil
	a.tail = n
	a.length = size
}

func (a *ArrayList[T]) Resize(size int) {
	if a.length == size {
		return
	}
	if a.length < size {
		a.upsize(size)
		return
	}
	a.downsize(size)
}

// Len returns the number of elemelnts in the array list
func (a *ArrayList[T]) Len() int {
	return a.length
}

type node[T any] struct {
	data [nodeLength]T
	next *node[T]
}

func (a *node[T]) forward(i int) (*node[T], int) {
	n := a
	for i >= nodeLength {
		n = n.next
		i -= nodeLength
	}
	return n, i
}
