package actor

import "sync"

type World interface {
	Wait()
}

const annon PID = "annon"

func New(init Actor) World {
	w := &world{}
	inbox := make(chan *msg)
	go runWorld(w, inbox)
	inbox <- &msg{to: annon, from: annon, data: spawnActor{
		PID:   nextPID(),
		Actor: init,
	}}
	return w
}

func runWorld(w *world, inbox chan *msg) {
	mm := mailman{
		processes: map[PID]*process{},
		unread:    []*msg{},
		inbox:     inbox,
	}
	for message := range inbox {
		switch x := message.data.(type) {
		case spawnActor:
			proc := &process{
				pid:   x.PID,
				in:    make(chan *msg),
				actor: x.Actor,
			}
			mm.processes[proc.pid] = proc
			go proc.supervise(Ctx{&actorContext{router: inbox, process: proc}})
			w.alive.Add(1)
		case signal:
			switch x {
			case sigterm:
				if proc, ok := mm.processes[message.to]; ok {
					close(proc.in)
					delete(mm.processes, message.to)
					return
				}
			}

		default:
			mm.route(message)
		}
	}

}

type world struct {
	alive sync.WaitGroup
	inbox chan *msg
}

type spawnActor struct {
	PID
	Actor
}

func (w *world) Wait() {
	w.alive.Wait()
}
