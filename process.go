package actor

import (
	"fmt"
	"log/slog"
)

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

func (s signal) String() string {
	switch s {
	case sigterm:
		return "sigterm"
	case sigSupervise:
		return "sigsupervise"
	default:
		return "invalid"
	}
}

func (p *process) supervise(ctx Ctx) {
	var supervisors []PID
	scope := slog.With("scope", "supervise", "pid", p.pid)

	notifyErr := func(err error) {
		for _, s := range supervisors {
			Send(ctx, s, Down{PID: p.pid, Error: err})
		}
	}

	defer scope.Info("closed")
	defer func() {
		if err := recover(); err != nil {
			scope.With("panic", err).Error("Recoverd panic")
			notifyErr(fmt.Errorf("%w recovered panic: %+v", ErrPanic, err))
		}
	}()

	for m := range p.in {
		m.slog(scope).Info("process recieved message")
		switch h := m.data.(type) {
		case addSupervisor:
			supervisors = append(supervisors, h.PID)
			continue
		}
		err := p.actor(ctx, m.from, m.data)
		if err != nil {
			notifyErr(err)
			return
		}
	}
}
