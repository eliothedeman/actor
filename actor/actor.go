package actor

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

var ErrDie = errors.New("die")

// type Proto = func(C)

// func New(p Proto) *World {
// 	w := World{
// 		ctx: ctx{
// 			h:     map[reflect.Type]func(C, PID, any) error{},
// 			alive: sync.WaitGroup{},
// 		},
// 	}

// 	Spawn(&w.ctx, p)
// 	return &w
// }

// type ctx struct {
// 	h     handlers
// 	alive sync.WaitGroup
// }

// type World struct {
// 	ctx
// }

// func (c *ctx) Wait() {

// }

// func (c *ctx) handlers() handlers {
// 	return c.h
// }

// type C interface {
// 	handlers() handlers
// 	Wait()
// }

// func Discard[T Msg](C, PID, T) error {
// 	return nil
// }

type PID int
type Down struct {
	Err error
}
type Unit struct{}

// type Scalar interface {
// 	constraints.Integer | constraints.Float | string | bool | Down | time.Time | ~struct{ comparable }
// }

// type IntSlices interface {
// 	~[]int | ~[]int8 | ~[]int16 | ~[]int32 | ~[]int64 | ~[]uint | ~[]uint8 | ~[]uint16 | ~[]uint32 | ~[]uint64
// }

// type FloatSlices interface {
// 	~[]float32 | ~[]float64
// }
// type StringSlices interface {
// 	~[]string
// }
// type Slices interface {
// 	IntSlices | FloatSlices | StringSlices
// }

// type IntMaps interface {
// 	~map[string]int | ~map[string]int8 | ~map[string]int16 | ~map[string]int32 | ~map[string]int64 | ~map[string]uint | ~map[string]uint8 | ~map[string]uint16 | ~map[string]uint32 | ~map[string]uint64
// }

// type FloatMaps interface {
// 	~map[string]float32 | ~map[string]float64
// }

// type Maps interface {
// 	IntMaps | FloatMaps | map[string]string | map[string]PID
// }

type pidMgr struct{}

func (p *pidMgr) next() PID {
	return 0
}

func (p *pidMgr) get(PID) *proc {
	return nil
}

func (p *pidMgr) set(PID, *proc) {

}

func (p *pidMgr) add(*proc) PID {
	return p.next()
}

type Msg interface {
	comparable
}
type actor func(C, PID, any)
type A[T Msg] func(C, PID, T)
type Selfer interface {
	Self() PID
}
type C interface {
	Selfer
	Sender
	Spawner
	Manager
}
type Binder interface {
	bind(reflect.Type, actor)
	call(C, PID, any)
}

func B[T Msg](b Binder, f func(C, PID, T)) {
	var t T
	b.bind(reflect.TypeOf(t), func(c C, p PID, a any) {
		f(c, p, a.(T))
	})
}

type Manager interface {
	manage(PID)
}

func Manage(m Manager, p PID) PID {
	return 0
}

type Spawner interface {
	binder() Binder
	spawn(Binder) PID
}

func Spawn(s Spawner, f func(Binder)) PID {
	b := s.binder()
	f(b)
	return s.spawn(b)
}

type MB interface {
	Manager
	Binder
	Spawner
}

func MSpawn(m MB, f func(Binder)) PID {
	return Manage(m, Spawn(m, f))
}

type Sender interface {
	send(to PID, msg any)
}

func Send[T Msg](s Sender, to PID, msg T) {
	s.send(to, msg)
}

func Exit() {
	panic(ErrDie)
}

func SendAt[T Msg](s Sender, to PID, msg T, at time.Time) {
	go func() {
		now := time.Now()
		if now.After(at) {
			Send(s, to, msg)
			return
		}
		time.Sleep(at.Sub(now))
		Send(s, to, msg)
	}()
}

func Discard[T Msg](C, PID, T) {}

type Waiter interface {
	Wait()
}

type proc struct {
	parent *proc
	PID
	wg       sync.WaitGroup
	handlers map[reflect.Type]actor
	children []*proc
}

// bind implements Binder.
func (p *proc) bind(key reflect.Type, a actor) {
	p.handlers[key] = a
}

// call implements Binder.
func (p *proc) call(c C, pid PID, a any) {
	t := reflect.TypeOf(a)
	p.handlers[t](c, pid, a)
}

// Wait implements Waiter.
func (p *proc) Wait() {
	p.wg.Wait()
}

// Self implements C.
func (p *proc) Self() PID {
	return p.PID
}

// binder implements C.
func (p *proc) binder() Binder {
	return p
}

// manage implements C.
func (p *proc) manage(pid PID) {
	panic("unimplemented")
}

// send implements C.
func (p *proc) send(to PID, msg any) {
	panic("unimplemented")
}

// spawn implements C.
func (p *proc) spawn(b Binder) PID {
	child := &proc{
		parent:   p,
		PID:      0,
		wg:       sync.WaitGroup{},
		handlers: map[reflect.Type]actor{},
		children: []*proc{},
	}
	p.children = append(p.children, child)
	panic("unimplemented")
}

func newProc() *proc {
	return &proc{}
}

func New(f func(Binder)) Waiter {
	b := newProc()
	Spawn(b, f)
	return b
}

// type handlers map[reflect.Type]func(C, PID, any) error

// type Handle[T Msg] func(self C, from PID, msg T) error

// func Bind[T Msg](c C, f Handle[T]) {
// 	h := c.handlers()
// 	var t T
// 	h[reflect.TypeOf(t)] = func(c C, p PID, a any) error {
// 		return f(c, p, a.(T))
// 	}
// }

// func Spawn(c C, prototype func(C)) PID {
// 	return 0
// }

// func MSpawn(c C, prototype func(C)) PID {
// 	// todo add monitoring here
// 	return Spawn(c, prototype)
// }

// func Send[T Msg](c C, to PID, msg T) {

// }
