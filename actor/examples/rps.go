package main

import (
	"log"
	"math/rand"
	"time"

	. "github.com/eliothedeman/actor/actor"
)

type throw int

const (
	rock throw = iota
	paper
	scisors
)

type win int
type lose int

func player(b Binder) {
	wins := 0
	B(b, func(self C, from PID, deadline time.Time) {
		Send(self, from, throw(rand.Int()%3))
	})
	B(b, func(self C, from PID, msg win) {
		wins++
	})
	B(b, func(self C, from PID, msg lose) {
		Send(self, from, wins)
		Exit()
	})
}

type game struct {
	players []PID
}

func (g *game) bind(c Binder) {
	B(c, func(c C, p PID, s start) {
		for i := 0; i < 10; i++ {
			g.players = append(
				g.players,
				Spawn(c, player),
			)
		}
	})
	// let players die
	B[Down](c, Discard)
	B(c, func(self C, from PID, t throw) {
	})
	complete := 0
	B(c, func(self C, from PID, wins int) {
		complete += 1
		log.Printf("player %d won %d times", from, wins)
		if complete >= 10 {
			Exit()
		}
	})
}

type start struct{}

func main() {
	New(new(game).bind).Wait()
}
