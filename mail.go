package actor

type msg struct {
	data any
	from PID
	to   PID
}

type mailman struct {
	processes map[PID]*process
	unread    []*msg
	inbox     chan *msg
}

func (m *mailman) route(message *msg) {
	if proc, ok := m.processes[message.to]; ok {
		proc.in <- message
		return
	}

	m.unread = append(m.unread, message)
}

type actorContext struct {
	router chan *msg
	*process
}

func (a *actorContext) PID() PID {
	return a.pid
}
