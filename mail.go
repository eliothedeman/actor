package actor

import "github.com/eliothedeman/actor/queue"

type msg struct {
	data any
	from PID
	to   PID
}

type mailman struct {
	processes map[PID]*process
	inbox     *queue.MQueue[*msg]
}

func (m *mailman) route(message *msg) {
	if proc, ok := m.processes[message.to]; ok {
		proc.in <- message
		return
	}
}

type actorContext struct {
	*process
	*world
}

func (a *actorContext) PID() PID {
	return a.pid
}
