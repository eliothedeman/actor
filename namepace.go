package actor

import (
	"fmt"
	"log"
	"runtime/debug"
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
	s := new(supervisor)
	s.init(PID{
		ns: i, actor: 0,
	})
	nx := &namespace{
		id: i,
		actors: map[id]Actor{
			0: s,
		},
	}

	n.local[i] = nx
	go nx.loop()
	return nx.spawn(a)
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
	actors      map[id]Actor
	activeActor id
}

func (n *namespace) spawn(a Actor) PID {
	i := id(0)
	for {
		i++
		if _, exists := n.actors[i]; !exists {
			pid := PID{ns: n.id, actor: i}
			a.init(pid)
			n.actors[i] = a
			return pid
		}
	}
}

func (n *namespace) handleDeadActor(actor id) {
	n.send(mail{
		PID: PID{ns: n.id, actor: 0},
		msg: ChildDie(PID{
			ns: n.id, actor: actor,
		}),
	})
}

func (n *namespace) loopOnce() (err error) {
	msg := n.recieve()
	defer func() {
		if e := recover(); e != nil {
			switch e := e.(type) {
			case errDie:
				n.handleDeadActor(msg.actor)
				// remove worker
			case error:
				debug.PrintStack()
				err = e
			default:
				debug.PrintStack()
				err = fmt.Errorf("paniced with non-error value %T:%+v", e, e)

			}
		}
		n.activeActor = 0
	}()

	// sent to the wrong namespace. Or potentially needs to go to another ndoe.
	if msg.ns != n.id {
		thisNode.send(msg)
		return
	}

	n.Lock()
	defer n.Unlock()
	n.activeActor = msg.actor

	if a, ok := n.actors[msg.actor]; ok {
		a.Recieve(msg.msg)
	} else {
		log.Panicf("Unable to route message %T:%+v", msg, msg)
	}
	return
}

func (n *namespace) loop() {
	for {
		err := n.loopOnce()
		if err != nil {
			panic(err)
		}
	}
}

type Handle interface {
	Send(msg any)
	Supervise(other Handle)
}

type handle struct {
	PID
}

func (h *handle) Supervise(other Handle) {
	thisNode.send(mail{
		PID: PID{
			ns:    other.(*handle).ns,
			actor: 0,
		},
		msg: startSupervise{
			parent: h.PID,
			child:  other.(*handle).PID,
		},
	})
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
