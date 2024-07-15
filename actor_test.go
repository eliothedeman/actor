package actor

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type wk struct {
	Ctx
	counter int
}

type newCount struct {
	reply PID
}

type crash struct{}

func count() Actor {
	i := 0
	return func(c Ctx, from PID, msg any) error {
		switch msg := msg.(type) {
		case int:
			i += msg
		case Down:
			MSpawn(c, count())
		case newCount:
			Send(c, msg.reply, Spawn(c, count()))
		case crash:
			panic("test")

		case PID:
			Send(c, msg, i)
		default:
			log.Panicf("bad message %+v", msg)
			// wtf
		}
		return nil
	}
}

func TestSpawn(t *testing.T) {
	w := New(func(c Ctx, from PID, message any) error {
		switch msg := message.(type) {
		case Init:
			pid := MSpawn(c, count())
			Send(c, pid, 100)
			Send(c, pid, 100)
			Send(c, pid, c.PID())
			Stop(c, pid)

		case int:
			assert.Equal(t, msg, 200)
			Stop(c, c.PID())
		}
		return nil
	})

	w.Wait()
}

func TestSupervise(t *testing.T) {
	New(func(c Ctx, from PID, message any) error {
		switch msg := message.(type) {
		case Init:
			pid := MSpawn(c, count())
			Send(c, pid, crash{})
		case Down:
			assert.ErrorIs(t, msg.Error, ErrPanic)
			StopSelf(c)
		}
		return nil
	}).Wait()
}
