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
	pid         PID
	in          chan *msg
	actor       func(Ctx, PID, any) error
	localMemory map[string]any
	binder
}

type Down struct {
	PID
	Err error
}

type unhanldedError struct {
	PID
	err error
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
	supervisor := PID("init")
	scope := slog.With("scope", "supervise", "pid", p.pid)
	if p.localMemory == nil {
		p.localMemory = make(map[string]any)
	}

	notifyErr := func(err error) {
		scope.Error("sending down signal")
		Send(ctx, supervisor, Down{PID: ctx.pid, Err: err})
		StopSelf(ctx)
	}

	defer scope.Info("closed")
	defer func() {
		if err := recover(); err != nil {
			scope.With("panic", err).Error("Recoverd panic")
			if err, ok := err.(error); ok {
				notifyErr(fmt.Errorf("%w recovered panic: %w", ErrPanic, err))
				return
			}
			notifyErr(fmt.Errorf("%w recovered panic: %+v", ErrPanic, err))
		}
	}()

	Bind[signal](ctx, func(c Ctx, from PID, message signal) error {
		switch message {
		case sigSupervise:
			supervisor = from
			scope = scope.With("supervisor", supervisor)
			return nil
		}
		return fmt.Errorf("unknown signal %s", message)
	})
	Bind[Down](ctx, func(c Ctx, from PID, message Down) error {
		return fmt.Errorf("unhandled error from %s %w", from, message.Err)
	})

	for m := range p.in {
		if err, found := p.binder.boundActor(ctx, m); found {
			if err != nil {
				notifyErr(err)
			}
			continue
		}
		m.slog(scope).Info("process recieved message")
		err := p.actor(ctx, m.from, m.data)
		if err != nil {
			notifyErr(err)
		}
	}
}
