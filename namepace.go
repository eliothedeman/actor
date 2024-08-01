package actor

import (
	"log"
	"sync"
	"sync/atomic"
)

type id uint32

type PID struct {
	ns    id
	actor id
}

var (
	namespaces atomic.Uint32
	thisNode   = &node{
		local: map[id]*namespace{},
	}
)

type node struct {
	local map[id]*namespace
	sync.Mutex
}

func (n *node) fork(a Actor) PID {
	n.Lock()
	defer n.Unlock()
	i := id(namespaces.Add(1))
	nx := &namespace{
		id: i,
		actors: map[id]Actor{
			0: a,
		},
	}
	p := PID{ns: i, actor: 0}

	a.init(p)
	n.local[i] = nx
	go nx.loop()
	return p
}

func (n *node) ns(i id) *namespace {
	n.Lock()
	defer n.Unlock()
	return n.local[i]
}

func (n *node) send(m mail) {
	n.local[m.ns].send(m)
}

type mail struct {
	PID
	msg any
}

type namespace struct {
	sync.Mutex
	id
	mailbox[mail]
	actors map[id]Actor
}

func (n *namespace) spawn(a Actor) PID {
	n.Lock()
	defer n.Unlock()
	i := id(0)
	for {
		i++
		if _, exists := n.actors[i]; !exists {
			pid := PID{ns: n.id, actor: i}
			n.actors[i] = a
			return pid
		}
	}

}

type nsState int

const (
	nsAgain nsState = iota
	nsStop
)

func (n *namespace) loopOnce() nsState {
	msg := n.recieve()

	// sent to the wrong namespace. Or potentially needs to go to another ndoe.
	if msg.ns != n.id {
		thisNode.send(msg)
		return nsAgain
	}

	if a, ok := n.actors[msg.actor]; ok {
		a.Recieve(msg.msg)
	} else {
		log.Fatalf("Unable to route message %T:%+v", msg, msg)
	}
	return nsAgain
}

func (n *namespace) loop() {
	for {
		switch n.loopOnce() {
		case nsAgain:
			continue
		case nsStop:
			return
		}
	}
}

type Handle interface {
	Send(msg any)
}

type handle struct {
	PID
}

func (h *handle) Send(msg any) {
	thisNode.send(mail{
		PID: h.PID, msg: msg,
	})
}

type Ctx struct {
	PID
}

func (c *Ctx) init(args PID) {
	c.PID = args
}
func (c *Ctx) Self() Handle {
	return &handle{c.PID}
}

func (c *Ctx) Spawn(other Actor) Handle {
	pid := thisNode.ns(c.ns).spawn(other)
	return &handle{pid}
}

type Actor interface {
	init(PID)
	Self() Handle
	Spawn(Actor) Handle
	Recieve(msg any)
}

func Fork(a Actor) Handle {
	return &handle{thisNode.fork(a)}
}

// func Spawn[T any](a Actor[T]) PID {
// 	return PID{}
// }

// func Send[T any](to PID, msg T) {

// }
