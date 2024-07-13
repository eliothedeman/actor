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
	return func(c *Ctx, msg any, from Addr) error {
		switch msg := msg.(type) {
		case int:
			i += msg
		case error:
			switch msg {
			case Death:
				NSpawn(c, count(), from)
			default:
				return msg
			}
		case Addr:
			Send(c, i, msg)
		default:
			// wtf
		}
		return nil
	}
}

func TestSpawn(t *testing.T) {
	w := New()
	c1 := Spawn(w, count())
	recieved := false
	c2 := Spawn(w, func(c *Ctx, msg any, from Addr) error {
		switch msg := msg.(type) {
		case int:
			assert.Equal(t, msg, 100)
			recieved = true
		}
		return nil
	})

	Send(w, 100, c1)
	Send(w, c2, c1)

	Send(w, Terminate, c1)
	Send(w, Terminate, c2)
	w.Wait()
	assert.True(t, recieved)
}
