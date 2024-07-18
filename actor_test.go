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
		case Init:
		default:
			log.Panicf("bad message %T %+v", msg, msg)
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

		case int:
			assert.Equal(t, msg, 200)
			Stop(c, from)
			StopSelf(c)
		}
		return nil
	})

	w.Wait()
}

func TestSupervise(t *testing.T) {
	New(func(c Ctx, from PID, message any) error {
		Bind[Down](c, func(c Ctx, from PID, message Down) error {
			assert.ErrorIs(t, message.Err, ErrPanic)
			StopSelf(c)
			return nil
		})
		pid := MSpawn(c, count())
		Send(c, pid, crash{})
		return nil
	}).Wait()
}

func TestKV(t *testing.T) {
	called := false
	New(func(c Ctx, from PID, message any) error {
		switch msg := message.(type) {
		case Init:
			Set(c, "tk", 400)
			Send(c, c.PID(), 400)
		case int:
			assert.Equal(t,
				msg,
				400,
			)
			assert.Equal(t, Get[int](c, "tk"), msg)
			StopSelf(c)
		}
		called = true
		return nil
	}).Wait()
	assert.True(t, called)
}

func doubleBind(c Ctx, from PID, m any) error {
	Bind[int](c, func(c Ctx, from PID, message int) error { return nil })
	Bind[int](c, func(c Ctx, from PID, message int) error { return nil })
	return nil
}

func TestBindTwice(t *testing.T) {
	called := false
	New(func(c Ctx, from PID, message any) error {
		Bind[Down](c, func(c Ctx, from PID, message Down) error {
			called = true
			assert.ErrorIs(t, message.Err, errDoubleBind)
			StopSelf(c)
			return nil
		})
		MSpawn(c, doubleBind)
		return nil
	}).Wait()
	assert.True(t, called)
}
