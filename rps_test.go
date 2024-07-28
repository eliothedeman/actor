package actor_test

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/eliothedeman/actor"
)

type throw int

const (
	rock throw = iota
	paper
	scisors
)

var (
	playTable [3][3]result
)

func init() {
	for me := rock; me <= scisors; me++ {
		for against := rock; against <= scisors; against++ {
			switch me {
			case rock:
				switch against {
				case rock:
					playTable[me][against] = draws
				case paper:
					playTable[me][against] = loses
				case scisors:
					playTable[me][against] = wins
				}
			case paper:
				switch against {
				case rock:
					playTable[me][against] = wins
				case paper:
					playTable[me][against] = draws
				case scisors:
					playTable[me][against] = loses
				}
			case scisors:
				switch against {
				case rock:
					playTable[me][against] = loses
				case paper:
					playTable[me][against] = wins
				case scisors:
					playTable[me][against] = draws
				}

			}

		}
	}
}

type result int

const (
	wins = iota
	loses
	draws
)

func (t throw) play(against throw) result {
	return playTable[t][against]
}

type win struct{}
type lose struct{}
type draw struct{}

func player(b Binder) {
	wins := 0
	B(b, func(self C, deadline time.Time) {
		Send(
			self,
			self.Parent(),
			RTS[throw]{From: self.Self(), Msg: throw(rand.Int() % 3)},
		)
	})
	B(b, func(self C, msg win) {
		wins++
	})
	B(b, func(self C, msg draw) {
	})
	B(b, func(self C, msg lose) {
		Send(self, self.Parent(), RTS[int]{self.Self(), wins})
		Exit()
	})
}

type game struct {
	playerCount int
	players     []PID
	throws      map[PID]throw
}

func (g *game) mostCommon() throw {
	counts := [scisors + 1]int{}
	for _, v := range g.throws {
		counts[v]++
	}
	maxVal := max(counts[0], counts[1], counts[2])
	for i := rock; i <= scisors; i++ {
		if counts[i] == maxVal {
			return i
		}
	}
	return scisors
}

func (g *game) playRound() (owins, odraws, oloses []PID) {
	against := g.mostCommon()
	for p, t := range g.throws {
		switch t.play(against) {
		case wins:
			owins = append(owins, p)
		case draws:
			odraws = append(odraws, p)
		case loses:
			oloses = append(oloses, p)
		}
	}
	return
}

func (g *game) bind(c Binder) {
	B(c, func(self C, _ Unit) {
		Send(self, self.Self(), start{})
	})
	B(c, func(c C, s start) {
		for i := 0; i < g.playerCount; i++ {
			g.players = append(
				g.players,
				Spawn(c, player),
			)
		}
		Alert(c, c.Self(), time.Now())
	})
	// let players die
	B[Down](c, Discard)
	B(c, func(self C, t RTS[throw]) {
		if g.throws == nil {
			g.throws = map[PID]throw{}
		}
		if _, exists := g.throws[t.From]; exists {
			Send(self, t.From, lose{})
			return
		}
		g.throws[t.From] = t.Msg
	})

	B(c, func(self C, t time.Time) {
		w, d, l := g.playRound()
		Broadcast(self, w, win{})
		Broadcast(self, l, lose{})
		Broadcast(self, d, draw{})
		nextGame := time.Now().Add(time.Nanosecond)
		g.throws = nil
		switch len(self.Children()) {
		case 1:
			Send(self, self.Children()[0], ErrDie)
		case 0:
			Exit()
		default:
			Broadcast(self, self.Children(), nextGame)
		}
		Alert(self, self.Self(), nextGame)
	})

	B(c, func(self C, wins RTS[int]) {
	})
}

type start struct{}

func TestRun(t *testing.T) {
	g := game{playerCount: 10000}
	New(g.bind).Wait()
}

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g := game{playerCount: 100}
		New(g.bind).Wait()
	}
}
