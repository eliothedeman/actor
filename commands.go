package actor

func Send(c Ctx, to PID, message any) {
	c.router <- c.msg(to, message)
}

func Spawn(c Ctx, a Actor) PID {
	pid := nextPID()
	c.router <- c.msg(annon, spawnActor{pid, a})
	return pid
}

type addSupervisor struct {
	PID
}

func Monitor(c Ctx, pid PID) {
	Send(c, pid, addSupervisor{})
}

func MSpawn(c Ctx, a Actor) PID {
	pid := Spawn(c, a)
	Monitor(c, pid)
	return pid
}

func Stop(c Ctx, pid PID) {
	Send(c, pid, sigterm)
}
