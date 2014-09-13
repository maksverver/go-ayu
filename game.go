package ayu

import "fmt"
import "io"

const S = 11  // board size

// The next player is either +1 (white) or -1 (black)
// Fields are represented as 0 (empty) or +1 (white piece) or -1 (black piece)
type State struct {
	Player int
	Fields [S][S]int
	Moves []Move
}

// Field coordinates locate a field on the board.
type Coords struct {
	Row,Col int
}

// A move consist of moving a piece from Src to Dst
type Move struct {
	Src,Dst Coords
}

func (coords Coords) String() string {
	return fmt.Sprintf("%c%d", 65+coords.Col, S-coords.Row)
}

func (m Move) String() string {
	return fmt.Sprintf("%s-%s", m.Src.String(), m.Dst.String())
}

func CreateState() *State {
	// TODO: initialize
	return &State{}
}

func ParseMove(s string) (m Move, ok bool) {
	// TODO!
	return
}

func (s *State) Over() bool {
	// TODO
	return true
}

func (s *State) Next() int {
	switch (s.Player) {
	case +1: return 0
	case -1: return 1
	}
	return -1
}

func (s *State) Valid(move Move) bool {
	// TODO
	return false
}

func (s *State) ListMoves() (moves []interface{}) {
	return
}

func (s *State) Execute(arg interface{}) bool {
	if m, ok := arg.(Move); ok && s.Valid(m) {
		s.Moves = append(s.Moves, m)
		// TODO: update board
		return true
	}
	return false
}

func (s *State) Scores() (int,int) {
	// TODO
	return 0,0
}

func (s *State) WriteLog (w io.Writer) {
	// TODO
}
