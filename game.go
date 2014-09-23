package ayu

import "fmt"
import "io"
import "regexp"
import "strconv"

const S = 11 // board size

var coords_re = regexp.MustCompile("^([A-K])([1-9]|1[01])$")
var move_re = regexp.MustCompile("^([A-K]([1-9]|1[01]))-([A-K]([1-9]|1[01]))$")

// Fields are represented as 0 (empty) or +1 (white piece) or -1 (black piece)
type Fields [S][S]int

// Game state consists of the current board and the history of moves played.
type State struct {
	Fields  Fields
	History []Move
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

func (c Coords) inRange() bool {
	return 0 <= c[0] && c[0] < S && 0 <= c[1] && c[1] < S
}

func abs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}

func (c Coords) distanceTo(d Coords) int {
	return abs(c[0] - d[0]) + abs(c[1] - d[1])
}

func (m Move) inRange() bool {
	return m[0].inRange() && m[1].inRange()
}

func (f *Fields) get(c Coords) *int {
	return &f[c[0]][c[1]]
}

func (f *Fields) set(c Coords, p int) {
	f[c[0]][c[1]] = p
}

func swapInts(p, q *int) {
	*p, *q = *q, *p
}

func (f *Fields) swap(x, y Coords) {
	swapFields(f.get(x), f.get(y))
}

func (f *Fields) relabelConnected(c Coords, p int, q int) int {
	res := 0
	if c.inRange() && *f.get(c) == p {
		f.set(c, q)
		res += 1 +
			f.relabelConnected(Coords{c[0] - 1, c[1]}, p, q) +
			f.relabelConnected(Coords{c[0] + 1, c[1]}, p, q) +
			f.relabelConnected(Coords{c[0], c[1] - 1}, p, q) +
			f.relabelConnected(Coords{c[0], c[1] + 1}, p, q)
	}
	return res
}

// Determines whether the given move keeps its unit connected.
func (f Fields) keepsUnitConnected(move Move) bool {
	player := *f.get(move[0])
	size := (&f).relabelConnected(move[0], player, 7)
	if size == 1 {
		return move[0].distanceTo(move[1]) <= 1
	}
	f.swap(move[0], move[1])
	return (&f).relabelConnected(move[1], 7, player) == size
}

// Calculates the distance to the nearest friendly unit from the unit
// indicated by the given coordinates.  Returns 0 if no friendly units
// are reachable.
func (f *Fields) distanceToNearestFriendlyUnit(c Coords) int {
	// TODO
	return 0
}

// Determines whether the given move reduces the distance to the nearest
// friendly unit.
func (f Fields) reducesDistanceToNearestFriendlyUnit(m Move) bool {
	d := f.distanceToNearestFriendlyUnit(m[0])
	if d <= 0 {
		return false
	}
	f.swap(m[0], m[1])
	return f.distanceToNearestFriendlyUnit(m[1]) < d
}

func (s *State) Valid(m Move) bool {
	p := s.NextPlayer()
	return m.inRange() &&
		*s.Fields.get(m[0]) == p &&
		*s.Fields.get(m[1]) == 0 &&
		s.Fields.keepsUnitConnected(m) // &&
	// s.Fields.reducesDistanceToNearestFriendlyUnit(m)
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
		s.Fields.swap(m[0], m[1])
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

func (fields *Fields) WriteBoard(w io.Writer) {
	var line [S + 1]byte
	line[S] = '\n'
	for i := 0; i < S; i++ {
		for j := 0; j < S; j++ {
			switch fields[i][j] {
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

// TODO: this should return an error
func (s *State) WriteLog(w io.Writer) {
	for i := 0; i < len(s.History)/2; i++ {
		fmt.Fprintf(w, "%3d. %-8s %s\n", i+1, s.History[2*i], s.History[2*i+1])
	}
	if len(s.History)%2 == 1 {
		fmt.Fprintf(w, "%3d. %s\n", len(s.History)/2+1, s.History[len(s.History)-1])
	}
}
