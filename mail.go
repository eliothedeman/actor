package actor

import "sync"

type Addr string
type msg struct {
	data any
	from Addr
}

type World interface {
	Spawn(Addr, Actor)
	Wait()
}

func New() World {
	return &world{}
}

type world struct {
	mm mailman
}

func (w *world) Spawn(at Addr, a Actor) {
	w.mm.spawn(at, a)
	w.mm.send(at, "", Signal{false, Init})
}

func (w *world) Wait() {
	w.mm.alive.Wait()
}

type mailman struct {
	processes map[PID]*process
	mailboxes map[Addr]chan *msg
	unread    []*msg
	alive     sync.WaitGroup
	sync.Mutex
}

func (m *mailman) ensureInit() {
	if m.processes == nil {
		m.processes = make(map[PID]*process)
	}
	if m.mailboxes == nil {
		m.mailboxes = make(map[Addr]chan *msg)
	}
}

func (m *mailman) spawn(at Addr, a Actor) PID {
	m.Lock()
	defer m.Unlock()
	m.ensureInit()

	// actor is already running, return the existing pid
	if c, alreadyRunning := m.mailboxes[at]; alreadyRunning {
		for _, proc := range m.processes {
			if proc.in == c {
				return proc.pid
			}
		}
	}
	m.alive.Add(1)
	proc := new(process)
	proc.actor = a
	proc.in = make(chan *msg)
	proc.pid = nextPID()
	proc.addr = at

	m.mailboxes[at] = proc.in
	m.processes[proc.pid] = proc
	// todo implement public ctx
	go proc.supervise(&actorContext{mailman: m, process: proc})
	return proc.pid
}

func (m *mailman) send(to Addr, from Addr, data any) {
	m.Lock()
	defer m.Unlock()
	m.ensureInit()
	if box, ok := m.mailboxes[to]; ok {
		box <- &msg{data: data, from: from}
	}

	m.unread = append(m.unread, &msg{data: data, from: from})
}

func (m *mailman) monitor(parent Addr, child PID) {
	m.Lock()
	proc, exists := m.processes[child]
	m.Unlock()
	if !exists {
		// TODO send an error to the parent
	}
	m.send(proc.addr, parent, &addSupervisor{parent})
}

type actorContext struct {
	*mailman
	*process
}

func (a *actorContext) Self() Addr {
	return a.process.addr
}

func (a *actorContext) Send(to Addr, message any) {
	a.mailman.send(to, a.Self(), message)
}

func (a *actorContext) Spawn(at Addr, actor Actor) PID {
	return a.mailman.spawn(at, actor)
}

func (a *actorContext) Monitor(pid PID) {
	a.mailman.monitor(a.Self(), pid)
}
