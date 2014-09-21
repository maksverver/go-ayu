package ayu

import "fmt"
import "io"
import "regexp"
import "strconv"

const S = 11 // board size

var coords_re = regexp.MustCompile("^([A-K])([1-9]|1[01])$")
var move_re = regexp.MustCompile("^([A-K]([1-9]|1[01]))-([A-K]([1-9]|1[01]))$")

// The next player is either +1 (white) or -1 (black)
// Fields are represented as 0 (empty) or +1 (white piece) or -1 (black piece)
type State struct {
	Fields  [S][S]int // current board state
	History []Move    // list of moves played so far
}

// Field coordinates locate a field on the board.
type Coords [2]int // [row, column]

// A move consist of moving a piece at a source field to a destination field.
type Move [2]Coords // [source, destination]

func (coords Coords) String() string {
	return fmt.Sprintf("%c%d", 65+coords[1], S-coords[0])
}

func (m Move) String() string {
	return fmt.Sprintf("%s-%s", m[0].String(), m[1].String())
}

func CreateState() *State {
	var res State
	for i := 0; i < S; i++ {
		for j := 0; j < S; j++ {
			res.Fields[i][j] = j%2 - i%2
		}
	}
	res.History = make([]Move, 0)
	return &res
}

func ParseCoords(s string) (c Coords, ok bool) {
	if matches := coords_re.FindStringSubmatch(s); matches != nil {
		// No need to check for parse errors.  Coords regexp only
		// accepts correctly formatted coordinates.
		x, _ := strconv.ParseInt(matches[1], 36, 0) // x in [10..20]
		y, _ := strconv.ParseInt(matches[2], 10, 0) // y in [1..11]
		c[0] = S - int(y)
		c[1] = int(x) - 10
		ok = true
	}
	return
}

func ParseMove(s string) (m Move, ok bool) {
	if matches := move_re.FindStringSubmatch(s); matches != nil {
		// No need to check result of ParseCoords.  Move regexp only
		// accepts correctly formatted coordinates.
		m[0], _ = ParseCoords(matches[1])
		m[1], _ = ParseCoords(matches[3])
		ok = true
	}
	return
}

// Note: Next() is called by the arbiter and return 0 (white) or 1 (black)
func (s *State) Next() int {
	return len(s.History) % 2
}

// Used internally; returns +1 (white) or -1 (black)
func (s *State) NextPlayer() int {
	return 1 - len(s.History)%2*2
}

func (s *State) generateMovesToChan(ch chan<- Move) {
	var move Move
	for move[0][0] = 0; move[0][0] < S; move[0][0]++ {
		for move[0][1] = 0; move[0][1] < S; move[0][1]++ {
			for move[1][0] = 0; move[1][0] < S; move[1][0]++ {
				for move[1][1] = 0; move[1][1] < S; move[1][1]++ {
					if s.Valid(move) {
						ch <- move
					}
				}
			}
		}
	}
	close(ch)
}

func (s *State) generateMoves() <-chan Move {
	ch := make(chan Move)
	go s.generateMovesToChan(ch)
	return ch
}

func coordsInRange(c Coords) bool {
	return 0 <= c[0] && c[0] < S && 0 <= c[1] && c[1] < S
}

func (s *State) Valid(move Move) bool {
	p := s.NextPlayer()
	// TODO: real implementation
	// Need to check two conditions:
	//  - move keeps unit connected
	//  - move must reduce distance to nearest other unit
	return coordsInRange(move[0]) && coordsInRange(move[1]) && s.Fields[move[0][0]][move[0][1]] == p && s.Fields[move[1][0]][move[1][1]] == 0
}

func (s *State) Over() bool {
	// The games is over iff. the next player has no possible moves.
	_, ok := <-s.generateMoves()
	return !ok
}

func (s *State) ListMoves() (moves []interface{}) {
	for move := range s.generateMoves() {
		moves = append(moves, move)
	}
	return
}

func (s *State) Execute(arg interface{}) bool {
	if m, ok := arg.(Move); ok && s.Valid(m) {
		s.History = append(s.History, m)
		p := &s.Fields[m[0][0]][m[0][1]]
		q := &s.Fields[m[1][0]][m[1][1]]
		*p, *q = *q, *p
		return true
	}
	return false
}

func (s *State) Scores() (int, int) {
	if s.Over() {
		if s.Next() == 0 {
			return 1, 0
		} else {
			return 0, 1
		}
	} else {
		return 0, 0
	}
}

func (s *State) WriteBoard(w io.Writer) {
	var line [S + 1]byte
	line[S] = '\n'
	for i := 0; i < S; i++ {
		for j := 0; j < S; j++ {
			switch s.Fields[i][j] {
			case 0:
				line[j] = '.'
			case -1:
				line[j] = '-'
			case +1:
				line[j] = '+'
			default:
				line[j] = '#'
			}
		}
		w.Write(line[:])
	}
}

func (s *State) WriteLog(w io.Writer) {
	// TODO
	w.Write([]byte("TODO"))
}
