package actor

import (
	"fmt"
	"log/slog"
	"sync"
)

type World interface {
	Wait()
}

const annon PID = "annon"

func New(init Actor) World {
	w := &world{}
	inbox := make(chan *msg)
	go runWorld(w, inbox)
	pid := nextPID()
	inbox <- &msg{to: annon, from: annon, data: spawnActor{
		PID:   pid,
		Actor: init,
	}}
	inbox <- &msg{to: pid, from: annon, data: Init{}}
	return w
}

func runWorld(w *world, inbox chan *msg) {
	mm := mailman{
		processes: map[PID]*process{},
		unread:    []*msg{},
		inbox:     inbox,
	}
	for message := range inbox {
		l := slog.With("to", message.to, "from", message.from)
		l.Info("world recieved", "type", fmt.Sprintf("%T", message.data), "data", message.data)
		switch x := message.data.(type) {
		case spawnActor:
			proc := &process{
				pid:   x.PID,
				in:    make(chan *msg),
				actor: x.Actor,
			}
			mm.processes[proc.pid] = proc
			go func() {
				defer w.alive.Done()
				proc.supervise(Ctx{&actorContext{router: inbox, process: proc}})
			}()
			w.alive.Add(1)
		case signal:
			l.Info("routing signal", "signal", x)
			switch x {
			case sigterm:
				if proc, ok := mm.processes[message.to]; ok {
					close(proc.in)
					delete(mm.processes, message.to)
					continue
				}
			}

		default:
			l.Info("routing message")
			mm.route(message)
		}
	}

}

type world struct {
	alive sync.WaitGroup
}

type spawnActor struct {
	PID
	Actor
}

func (w *world) Wait() {
	w.alive.Wait()
}
