package actor

import (
	"sync/atomic"
)

type errDie struct{}

func (errDie) Error() string {
	return "die"
}
func (errDie) String() string {
	return "die"
}

var ErrDie = errDie{}

type ChildDie PID

var nextPid atomic.Int64

// func next() PID {
// 	return PID(nextPid.Add(1))
// }

// type PID int

// func (p PID) Slog(l *slog.Logger) *slog.Logger {
// 	if l == nil {
// 		l = slog.Default()
// 	}
// 	return l.With("pid", p)
// }

// type Down struct {
// 	Err error
// }
// type Unit struct{}

// func (p *proc) mkChild() *proc {
// 	x := &proc{
// 		parent:   p,
// 		self:     next(),
// 		wg:       sync.WaitGroup{},
// 		handlers: map[reflect.Type]actor{},
// 		in:       newMailBox[any](),
// 	}
// 	p.children = append(p.children, x)
// 	return x
// }

// type Msg interface {
// 	comparable
// }
// type actor func(C, any)
// type A[T Msg] func(C, T)
// type Selfer interface {
// 	Self() PID
// }

// // returns parent pid or self if root
// type Parenter interface {
// 	Parent() PID
// }
// type C interface {
// 	Selfer
// 	Parenter
// 	Children() []PID
// 	Sender
// 	Spawner
// 	Manager
// }
// type Binder interface {
// 	bind(reflect.Type, actor)
// 	call(C, any)
// }

// func B[T Msg](b Binder, f func(C, T)) {
// 	var t T
// 	b.bind(reflect.TypeOf(t), func(c C, a any) {
// 		f(c, a.(T))
// 	})
// }

// type Manager interface {
// 	manage(PID)
// }

// func Manage(m Manager, p PID) PID {
// 	return 0
// }

// type Spawner interface {
// 	spawn(f func(Binder)) PID
// }

// func Spawn(s Spawner, f func(Binder)) PID {
// 	return s.spawn(f)
// }

// type MB interface {
// 	Manager
// 	Binder
// 	Spawner
// }

// func MSpawn(m MB, f func(Binder)) PID {
// 	return Manage(m, Spawn(m, f))
// }

// type Sender interface {
// 	send(to PID, msg any) bool
// }

// func Send[T Msg](s Sender, to PID, msg T) {
// 	s.send(to, msg)
// }

// func Exit() {
// 	panic(ErrDie)
// }

// func SendAt[T Msg](s Sender, to PID, msg T, at time.Time) {
// 	go func() {
// 		now := time.Now()
// 		if now.After(at) {
// 			Send(s, to, msg)
// 			return
// 		}
// 		time.Sleep(at.Sub(now))
// 		Send(s, to, msg)
// 	}()
// }
// func Broadcast(s Sender, to []PID, msg any) {
// 	for _, t := range to {
// 		Send(s, t, msg)
// 	}
// }

// func Alert(s Sender, to PID, at time.Time) {
// 	SendAt(s, to, at, at)
// }

// func Discard[T Msg](C, T) {}

// type Waiter interface {
// 	Wait()
// }

// func (p *proc) slog() *slog.Logger {
// 	return p.self.Slog(nil)
// }

// type proc struct {
// 	parent   *proc
// 	self     PID
// 	wg       sync.WaitGroup
// 	handlers map[reflect.Type]actor
// 	in       *mailbox[any]
// 	children []*proc
// }

// // bind implements Binder.
// func (p *proc) bind(key reflect.Type, a actor) {
// 	if p.handlers == nil {
// 		p.handlers = make(map[reflect.Type]actor)
// 	}
// 	p.handlers[key] = a
// }

// // call implements Binder.
// func (p *proc) call(c C, a any) {
// 	t := reflect.TypeOf(a)
// 	if fn, ok := p.handlers[t]; ok {
// 		fn(c, a)
// 		return
// 	}
// 	p.self.Slog(nil).
// 		WithGroup("call").
// 		With("payload", t, "handlers", p.handlers).
// 		Error("unhandled payload type")
// 	panic(ErrDie)
// }

// // Wait implements Waiter.
// func (p *proc) Wait() {
// 	p.wg.Wait()
// }

// // Self implements C.
// func (p *proc) Self() PID {
// 	return p.self
// }

// func (p *proc) Parent() PID {
// 	if p.parent == nil {
// 		return p.self
// 	}
// 	return p.parent.self
// }

// func (p *proc) Children() (out []PID) {
// 	for _, c := range p.children {
// 		out = append(out, c.self)
// 	}
// 	return
// }

// type RTS[T any] struct {
// 	From PID
// 	Msg  T
// }

// // manage implements C.
// func (p *proc) manage(pid PID) {
// 	panic("unimplemented manage")
// }

// // send implements C.
// func (p *proc) send(to PID, msg any) bool {
// 	// dead
// 	if p == nil {
// 		return false
// 	}
// 	// p.self.Slog(nil).
// 	// 	With("to", to, "msg", fmt.Sprintf("%T:%+v", msg, msg)).
// 	// 	Debug("sending")
// 	if p.self == to {
// 		p.in.send(msg)
// 		return true
// 	}

// 	if p.parent != nil && p.parent.self == to {
// 		p.parent.in.send(msg)
// 		return true
// 	}
// 	for _, c := range p.children {
// 		if c.send(to, msg) {
// 			return true
// 		}
// 	}
// 	return false
// }

// // spawn implements C.
// func (p *proc) spawn(f func(Binder)) PID {
// 	child := p.mkChild()
// 	// p.self.Slog(nil).Debug("spawning", "child", child.self, "actor", f)
// 	f(child)
// 	go child.run()
// 	return child.self
// }

// func (p *proc) runOnce(payload any) (err error) {
// 	defer func() {
// 		if mErr := recover(); mErr != nil {
// 			switch x := mErr.(type) {
// 			case error:
// 				if x != ErrDie {
// 					debug.PrintStack()
// 				}
// 				err = x
// 			default:
// 				err = fmt.Errorf("recovered from panic with unknown value %+v", err)
// 			}
// 		}
// 	}()
// 	switch x := payload.(type) {
// 	case errDie:
// 		return ErrDie
// 	case childDie:
// 		p.children = slices.DeleteFunc(p.children, func(c *proc) bool {
// 			return c.self == PID(x)
// 		})

// 	default:
// 		p.call(p, payload)
// 	}
// 	return nil
// }

// var allAlive = new(atomic.Int64)

// func (p *proc) run() {
// 	defer func() {
// 		p.wg.Done()
// 		if p.parent != nil {
// 			p.parent.in.send(childDie(p.self))
// 		}
// 		allAlive.Add(-1)
// 	}()
// 	allAlive.Add(1)
// 	p.wg.Add(1)
// 	for {
// 		m := p.in.recieve()
// 		err := p.runOnce(m)
// 		if err != nil {
// 			// p.Self().Slog(nil).Error("exiting", "err", err, "h", p.handlers)
// 			return
// 		}
// 	}
// }

// func New(f func(Binder)) Waiter {

// 	root := &proc{
// 		parent:   nil,
// 		self:     next(),
// 		wg:       sync.WaitGroup{},
// 		handlers: map[reflect.Type]actor{},
// 		in:       newMailBox[any](),
// 	}
// 	f(root)
// 	go Send(root, root.self, Unit{})
// 	root.run()
// 	return root
// }
