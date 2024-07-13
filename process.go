package actor

type PID string

type Ctx interface {
	Self() Addr
	Send(to Addr, message any)
	Spawn(Addr, func(Ctx, Addr, any) error) PID
	Monitor(PID)
}

type process struct {
	addr        Addr
	pid         PID
	in          chan *msg
	actor       Actor
	supervisors []Addr
}

type Code int

const (
	Alive Code = 1 << iota
	Down
	Init
)

type Signal struct {
	handled bool
	code    Code
}

func (s *Signal) Handle() Code {
	s.handled = true
	return s.code
}

type Err struct {
	Error error
	PID
	Addr
}

type addSupervisor struct {
	supervisor Addr
}

type UnhandledSignal struct {
	Addr
	Signal
}

func (p *process) supervise(ctx Ctx) {
	for m := range p.in {
		switch h := m.data.(type) {
		case *addSupervisor:
			p.supervisors = append(p.supervisors, h.supervisor)
			continue
		case Signal:
			err := p.actor(ctx, m.from, m.data)
			if err != nil {
				for _, s := range p.supervisors {
					ctx.Send(s, err)
				}
			}
			if !(h.handled) {
				for _, s := range p.supervisors {
					ctx.Send(s, &UnhandledSignal{ctx.Self(), h})
				}
			}
		}
		err := p.actor(ctx, m.from, m.data)
		if err != nil {
			for _, s := range p.supervisors {
				ctx.Send(s, err)
			}
		}
	}
}
