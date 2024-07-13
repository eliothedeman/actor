package actor

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
)

var (
	// Terminate indicates a process should not be restarted
	Terminate = errors.New("terminate")

	Death = errors.New("death")
)

var actorCounter atomic.Int64

func nextAddr() Addr {
	return Addr(strconv.Itoa(int(actorCounter.Add(1))))
}

type Addr string

func New() *Ctx {
	ctx := new(Ctx)
	ctx.children = make(actorSet)
	return ctx
}

type msg struct {
	data any
	from Addr
}

type Ctx struct {
	me       Addr
	parent   *Ctx
	actor    Actor
	children actorSet
	in       chan *msg
	wg       sync.WaitGroup
}

func (c *Ctx) Wait() {
	c.wg.Wait()
}

type actorSet map[Addr]*Ctx

func (ctx *Ctx) init(parent *Ctx, actor Actor) {
	// make surethereis only one thing running with this addr if it is arleady set
	ctx.parent = parent
	ctx.actor = actor
	ctx.me = nextAddr()
	ctx.children = make(actorSet)
	ctx.in = make(chan *msg)
	if parent == nil {
		return
	}
	parent.children[ctx.me] = ctx
}

func (c *Ctx) supervise(child *Ctx) {
	for m := range child.in {
		if m.data == Terminate {
			close(child.in)
			delete(c.children, child.me)
			c.wg.Done()
			return
		}

		err := child.actor(child, m.data, m.from)
		if err != nil {
			if err == Terminate {
				close(child.in)
				delete(c.children, child.me)
				c.wg.Done()
				return
			}
			c.in <- &msg{err, child.me}
		}
	}
}

func send(c *Ctx, data any, to Addr, from Addr) {
	if child, ok := c.children[to]; ok {
		child.in <- &msg{data, c.me}
		return
	}

	if c.parent == nil {
		log.Fatalf("Unable to find recipiant %s for %+v sent from %s", to, data, from)
	}
	send(c.parent, data, to, from)
}

func Send(c *Ctx, data any, to Addr) {
	send(c, data, to, c.me)
}

type Actor = func(c *Ctx, msg any, from Addr) error

func Spawn(c *Ctx, a Actor) Addr {
	child := new(Ctx)
	child.init(c, a)
	c.wg.Add(1)
	go c.supervise(child)
	return child.me
}

func NSpawn(c *Ctx, a Actor, name Addr) {
	child := new(Ctx)
	child.me = name
	child.init(c, a)
	go c.supervise(child)
}
