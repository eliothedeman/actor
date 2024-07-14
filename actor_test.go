package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type wk struct {
	Ctx
	counter int
}

type done struct{}

func count() Actor {
	i := 0
	return func(c Ctx, from PID, msg any) error {
		switch msg := msg.(type) {
		case int:
			i += msg
		case error:
			switch msg {
			case Death:
				MSpawn(c, count())
			default:
				return msg
			}
		case PID:
			Send(c, msg, i)
		default:
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
		}
		return nil
	})

	w.Wait()
	// jV
	// Spawn(w, count())
	// recieved := false
	// c2 := Spawn(w, func(c *Ctx, msg any, from Addr) error {
	// 	switch msg := msg.(type) {
	// 	case int:
	// 		assert.Equal(t, msg, 100)
	// 		recieved = true
	// 	}
	// 	return nil
	// })

	// Send(w, 100, c1)
	// Send(w, c2, c1)

	// Send(w, Terminate, c1)
	// Send(w, Terminate, c2)
	// w.Wait()
	// assert.True(t, recieved)
}
