package actor_test

import (
	"log"
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

func (t throw) String() string {
	switch t {
	case rock:
		return "rock"
	case paper:
		return "paper"
	case scisors:
		return "scisors"
	}
	return ""
}

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
	wins result = iota
	loses
	draws
)

func (r result) String() string {
	switch r {
	case wins:
		return "win"
	case loses:
		return "lose"
	case draws:
		return "draw"
	}
	return ""
}

func (t throw) play(against throw) result {
	return playTable[t][against]
}

type win struct{}
type lose struct{}
type draw struct{}

type playerThrow struct {
	player int
	throw  throw
}

// func player(b Binder) {
// 	wins := 0
// 	B(b, func(self C, deadline time.Time) {
// 		Send(
// 			self,
// 			self.Parent(),
// 			RTS[throw]{From: self.Self(), Msg: throw(rand.Int() % 3)},
// 		)
// 	})
// 	B(b, func(self C, msg win) {
// 		wins++
// 	})
// 	B(b, func(self C, msg draw) {
// 	})
// 	B(b, func(self C, msg lose) {
// 		Send(self, self.Parent(), RTS[int]{self.Self(), wins})
// 		Exit()
// 	})
// }

// func (g *game) playRound() (owins, odraws, oloses []PID) {
// 	against := g.mostCommon()
// 	for p, t := range g.throws {
// 		switch t.play(against) {
// 		case wins:
// 			owins = append(owins, p)
// 		case draws:
// 			odraws = append(odraws, p)
// 		case loses:
// 			oloses = append(oloses, p)
// 		}
// 	}
// 	return
// }

// func (g *game) bind(c Binder) {
// 	B(c, func(self C, _ Unit) {
// 		Send(self, self.Self(), start{})
// 	})
// 	B(c, func(c C, s start) {
// 		for i := 0; i < g.playerCount; i++ {
// 			g.players = append(
// 				g.players,
// 				Spawn(c, player),
// 			)
// 		}
// 		Alert(c, c.Self(), time.Now())
// 	})
// 	// let players die
// 	B[Down](c, Discard)
// 	B(c, func(self C, t RTS[throw]) {
// 		if g.throws == nil {
// 			g.throws = map[PID]throw{}
// 		}
// 		if _, exists := g.throws[t.From]; exists {
// 			Send(self, t.From, lose{})
// 			return
// 		}
// 		g.throws[t.From] = t.Msg
// 	})

// 	B(c, func(self C, t time.Time) {
// 		w, d, l := g.playRound()
// 		Broadcast(self, w, win{})
// 		Broadcast(self, l, lose{})
// 		Broadcast(self, d, draw{})
// 		nextGame := time.Now().Add(time.Nanosecond)
// 		g.throws = nil
// 		switch len(self.Children()) {
// 		case 1:
// 			Send(self, self.Children()[0], ErrDie)
// 		case 0:
// 			Exit()
// 		default:
// 			Broadcast(self, self.Children(), nextGame)
// 		}
// 		Alert(self, self.Self(), nextGame)
// 	})

// 	B(c, func(self C, wins RTS[int]) {
// 	})
// }

type start struct {
	done chan int
}

type player struct {
	Ctx
	id   int
	game Handle
	wins int
}

func badType(a any) {
	log.Fatalf("Bad type %T:%+v", a, a)
}

func (p *player) Recieve(msg any) {
	switch m := msg.(type) {
	case time.Time:
		p.game.Send(
			playerThrow{
				player: p.id,
				throw:  throw(rand.Int() % 3),
			},
		)
	case result:
		switch m {
		case wins:
			p.wins++
		case loses:
			p.game.Send(playerExit{p.id, p.wins})
		case draws:
		}
	default:
		badType(msg)
	}

}

type game struct {
	Ctx
	playerCount int
	players     map[int]Handle
	throws      []playerThrow
	done        chan int
}

func (g *game) mostCommon() throw {
	counts := [scisors + 1]int{}
	for _, v := range g.throws {
		counts[v.throw]++
	}
	maxVal := max(counts[0], counts[1], counts[2])
	for i := rock; i <= scisors; i++ {
		if counts[i] == maxVal {
			return i
		}
	}
	return scisors
}

type playerExit struct {
	player int
	wins   int
}

func (g *game) Recieve(msg any) {
	switch m := msg.(type) {
	case start:
		g.done = m.done
		g.players = make(map[int]Handle, g.playerCount)
		for i := range g.playerCount {
			g.players[i] = g.Spawn(&player{id: i, game: g.Self()})
		}
		g.Self().Send(time.Now())
	case playerThrow:
		g.throws = append(g.throws, m)

		// game time baby
	case time.Time:
		against := g.mostCommon()
		for _, t := range g.throws {
			res := t.throw.play(against)
			if _, ok := g.players[t.player]; ok {
				g.players[t.player].Send(res)
			}
		}
		g.throws = nil
		next := time.Now().Add(time.Millisecond)
		for _, p := range g.players {
			p.Send(next)
		}
		go func() {
			time.Sleep(time.Millisecond * 20)
			g.Self().Send(time.Now())
		}()

	case playerExit:
		delete(g.players, m.player)
		if len(g.players) == 1 {
			g.done <- m.wins
			return
		}
	default:
		badType(msg)

	}
}

func TestRun(t *testing.T) {
	g := &game{playerCount: 1000}
	done := make(chan int)
	Fork(g).Send(start{done})
	<-done
}

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g := &game{playerCount: 1000}
		done := make(chan int)
		Fork(g).Send(start{done})
		<-done
	}
}
