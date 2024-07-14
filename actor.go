package actor

import (
	"errors"
	"strconv"
	"sync/atomic"
)

var (
	// Terminate indicates a process should not be restarted
	Death = errors.New("death")
)

var actorCounter atomic.Int64
var processCounter atomic.Int64

func nextPID() PID {
	return PID(strconv.Itoa(int(processCounter.Add(1))))
}

type Actor = func(c Ctx, from PID, message any) error
