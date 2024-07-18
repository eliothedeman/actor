package actor

import (
	"log"
	"log/slog"
	"sync"

	"github.com/eliothedeman/actor/queue"
)

type World interface {
	Wait()
}

const annon PID = "annon"

func New(init Actor) World {
	w := &world{
		alive: sync.WaitGroup{},
		mailman: mailman{
			processes: map[PID]*process{},
			inbox:     queue.New[*msg](),
		},
		spawn: make(chan spawnActor, 1),
		stop:  make(chan PID),
		Mutex: sync.Mutex{},
		init:  init,
	}
	w.spawn <- spawnActor{
		PID:   nextPID(),
		Actor: init,
	}
	w.runOnce()
	go w.run()
	return w
}

const Root PID = "root"

func sendAnnon(inbox chan *msg, to PID, data any) {
	inbox <- &msg{to: to, from: annon, data: data}
}

func rootSupervisor(c Ctx, from PID, m any) error {
	switch m := m.(type) {
	case Down:
		log.Fatalf(m.Err.Error())
	}
	return nil
}

type mbox queue.MQueue[*msg]

type world struct {
	alive sync.WaitGroup
	mailman
	spawn chan spawnActor
	stop  chan PID
	init  func(Ctx, PID, any) error
	sync.Mutex
}

func (w *world) runOnce() {
	l := slog.With("scope", "world")
	select {
	case m := <-w.inbox.Out():
		m.slog(l).Info("recieved")
		w.route(m)
	case pid := <-w.stop:
		proc := w.processes[pid]
		if proc == nil {
			log.Fatalf("Attmping to stop already stopped pid %s", pid)
		}
		w.inbox.Filter(func(m *msg) bool { return m.to == pid })
		delete(w.processes, pid)
		close(proc.in)

	case s := <-w.spawn:
		w.alive.Add(1)
		proc := &process{
			pid:   s.PID,
			in:    make(chan *msg),
			actor: s.Actor,
		}
		w.processes[proc.pid] = proc
		w.inbox.Send(&msg{to: s.PID, from: s.PID, data: Init{}})
		ctx := Ctx{&actorContext{world: w, process: proc}}
		go func() {
			defer w.alive.Done()
			proc.supervise(ctx)
		}()
	}
}

func (w *world) run() {
	for {
		w.runOnce()
	}

}

type spawnActor struct {
	PID
	Actor func(Ctx, PID, any) error
}

func (w *world) Wait() {
	w.alive.Wait()
}
