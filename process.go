package actor

type PID string

type Ctx struct {
	*actorContext
}

func (c *Ctx) msg(to PID, data any) *msg {
	return &msg{
		to:   to,
		from: c.pid,
		data: data,
	}
}

type Init struct{}

type process struct {
	pid   PID
	in    chan *msg
	actor func(Ctx, PID, any) error
}

type signal int

const (
	sigterm signal = iota
	sigSupervise
)

func (p *process) supervise(ctx Ctx) {
	var supervisors []PID
	for m := range p.in {
		switch h := m.data.(type) {
		case addSupervisor:
			supervisors = append(supervisors, h.PID)
			continue
		}
		err := p.actor(ctx, m.from, m.data)
		if err != nil {
			for _, s := range supervisors {
				Send(ctx, s, err)
			}
		}
	}
}
