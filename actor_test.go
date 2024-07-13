package actor

import (
	"testing"
)

type wk struct {
	Ctx
	counter int
}

type done struct{}

func (w *wk) Recieve(msg any, from Addr) error {
	switch msg := msg.(type) {
	case int:
		w.counter += msg
	}
	return nil
}

type ptr interface {
	Actor
}

func count() Actor {
	i := 0
	return func(c Ctx, from Addr, msg any) error {
		switch msg := msg.(type) {
		case int:
			i += msg
		case error:
			switch msg {
			case Death:
				c.Spawn()
				MSpawn(c, from, count())
			default:
				return msg
			}
		case Addr:
			c.Send(msg, i)
		default:
			// wtf
		}
		return nil
	}
}

func TestSpawn(t *testing.T) {
	w := New()
	w.Spawn("root", func(c Ctx, from Addr, message any) error {
		switch msg := message.(type) {
		case *Signal:
			switch msg.Handle() {
			case Init:

			}

		}
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
