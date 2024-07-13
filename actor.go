package actor

import (
	"errors"
	"strconv"
	"sync/atomic"
)

var (
	// Terminate indicates a process should not be restarted
	Terminate = errors.New("terminate")
	Death     = errors.New("death")
)

var actorCounter atomic.Int64
var processCounter atomic.Int64

func nextAddr() Addr {
	return Addr(strconv.Itoa(int(actorCounter.Add(1))))
}
func nextPID() PID {
	return PID(strconv.Itoa(int(processCounter.Add(1))))
}

type Actor = func(c Ctx, from Addr, message any) error

func MSpawn(c Ctx, at Addr, a Actor) {
	pid := c.Spawn(at, a)
	c.Monitor(pid)
}
