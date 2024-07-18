package actor

func Send(c Ctx, to PID, message any) {
	c.mailman.inbox.Send(c.msg(to, message))
}

func Spawn(c Ctx, a Actor) PID {
	pid := nextPID()
	c.spawn <- spawnActor{pid, a}
	return pid
}

type addSupervisor struct {
	PID
}

func Monitor(c Ctx, pid PID) {
	Send(c, pid, sigSupervise)
}

func SendToSupervisor(c Ctx, data any) {
	Send(c, Root, data)
}

func MSpawn(c Ctx, a Actor) PID {
	pid := Spawn(c, a)
	Monitor(c, pid)
	return pid
}

func Stop(c Ctx, pid PID) {
	c.stop <- pid
}

func StopSelf(c Ctx) {
	Stop(c, c.PID())
}

// Set a value to be available in this process, even after it has been restarted.
func Set[T any](c Ctx, key string, value T) {
	c.localMemory[key] = value
}

func Get[T any](c Ctx, key string) T {
	return c.localMemory[key].(T)
}
