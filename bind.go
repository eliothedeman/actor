package actor

import (
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/exp/constraints"
)

type binder struct {
	handlers map[reflect.Type]Actor
}

func (b *binder) boundActor(c Ctx, m *msg) (error, bool) {
	mType := reflect.TypeOf(m.data)
	if h, ok := b.handlers[mType]; ok {
		return h(c, m.from, m.data), true
	}
	return nil, false
}

type Primatives interface {
	constraints.Integer | constraints.Float | ~string | Down
}

var errDoubleBind = errors.New("double bind")

func Bind[T Primatives](c Ctx, f func(c Ctx, from PID, message T) error) {
	b := &c.process.binder
	if b.handlers == nil {
		b.handlers = make(map[reflect.Type]Actor)
	}
	var t T

	mType := reflect.TypeOf(t)
	switch any(t).(type) {
	// overridable
	case Down:
	default:
		if _, ok := b.handlers[mType]; ok {
			panic(fmt.Errorf("%w: attempting to bind %s more than once", errDoubleBind, mType.String()))
		}
	}
	b.handlers[mType] = func(c Ctx, from PID, message any) error {
		return f(c, from, message.(T))
	}
}
