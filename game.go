package ayu

import "fmt"
import "io"
import "math"
import "regexp"
import "strconv"

const DefaultSize = 11

var coords_re = regexp.MustCompile("^([A-Z])([1-9][0-9]?)$")
var move_re = regexp.MustCompile("^([A-Z][1-9][0-9]?)-([A-Z][1-9][0-9]?)$")

// Fields are represented as 0 (empty) or +1 (white piece) or -1 (black piece)
type Fields [][]int

// The game history is simply the sequence of moves played so far.
type History []Move

// Game state consists of the current board and the history of moves played.
type State struct {
	Fields  Fields
	History History
}

// Field coordinates locate a field on the board.
type Coords [2]int // [row, column]

// A move consist of moving a piece at a source field to a destination field.
type Move [2]Coords // [source, destination]

func (coords Coords) String() string {
	return fmt.Sprintf("%c%d", 65+coords[1], 1+coords[0])
}

func (m Move) String() string {
	return fmt.Sprintf("%s-%s", m[0].String(), m[1].String())
}

func CreateState(size int) *State {
	var s State
	s.Create(size)
	return &s
}

func ParseCoords(s string) (c Coords, ok bool) {
	if matches := coords_re.FindStringSubmatch(s); matches != nil {
		// No need to check for parse errors.  Coords regexp only
		// accepts correctly formatted coordinates.
		x, _ := strconv.ParseInt(matches[1], 36, 0)
		y, _ := strconv.ParseInt(matches[2], 10, 0)
		c[0] = int(y) - 1
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
		m[1], _ = ParseCoords(matches[2])
		ok = true
	}
	return
}

func IsValidSize(size int) bool {
	return (size%2 == 1 && 3 <= size && size <= 19)
}

func (s *State) Create(size int) {
	if !IsValidSize(size) {
		panic(fmt.Sprintf("Invalid size: %d", size))
	}
	s.Fields = make([][]int, size)
	for i := 0; i < size; i++ {
		s.Fields[i] = make([]int, size)
		for j := 0; j < size; j++ {
			s.Fields[i][j] = j%2 - i%2
		}
	}
	s.History = make([]Move, 0)
}

// Note: Next() is called by the arbiter and return 0 (white) or 1 (black)
func (s *State) Next() int {
	return len(s.History) % 2
}

// Used internally; returns +1 (white) or -1 (black)
func (s *State) NextPlayer() int {
	return 1 - len(s.History)%2*2
}

func (s *State) generateMovesToChan(ch chan<- Move, n int) {
loop:
	for ; n > 0; n-- {
		for r1 := range s.Fields {
			for c1 := range s.Fields[r1] {
				for r2 := range s.Fields {
					for c2 := range s.Fields[r2] {
						move := Move{{r1, c1}, {r2, c2}}
						if s.Valid(move) {
							ch <- move
							continue loop
						}
					}
				}
			}
		}
		break
	}
	close(ch)
}

// Generates up to n moves.
func (s *State) generateMaxMoves(n int) <-chan Move {
	ch := make(chan Move)
	go s.generateMovesToChan(ch, n)
	return ch
}

func (s *State) generateAllMoves() <-chan Move {
	return s.generateMaxMoves(math.MaxInt32)
}

func (c Coords) inRange(f Fields) bool {
	return 0 <= c[0] && c[0] < len(f) && 0 <= c[1] && c[1] < len(f[c[0]])
}

func abs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}

func (c Coords) distanceTo(d Coords) int {
	return abs(c[0]-d[0]) + abs(c[1]-d[1])
}

func (c Coords) stepTo(dir int) Coords {
	switch dir {
	case 0:
		return Coords{c[0] + 1, c[1]}
	case 1:
		return Coords{c[0], c[1] + 1}
	case 2:
		return Coords{c[0] - 1, c[1]}
	case 3:
		return Coords{c[0], c[1] - 1}
	}
	panic(fmt.Sprintf("invalid direction: %d", dir))
}

func (m Move) inRange(f Fields) bool {
	return m[0].inRange(f) && m[1].inRange(f)
}

func (f Fields) get(c Coords) *int {
	return &f[c[0]][c[1]]
}

func (f Fields) set(c Coords, p int) {
	f[c[0]][c[1]] = p
}

func swapInts(p, q *int) {
	*p, *q = *q, *p
}

func (f Fields) swap(x, y Coords) {
	swapInts(f.get(x), f.get(y))
}

func (f Fields) relabelConnected(c Coords, p int, q int) int {
	res := 0
	if c.inRange(f) && *f.get(c) == p {
		f.set(c, q)
		res++
		for dir := 0; dir < 4; dir++ {
			res += f.relabelConnected(c.stepTo(dir), p, q)
		}
	}
	return res
}

func (f Fields) cloneZero() (g Fields) {
	g = make([][]int, len(f))
	for i := range f {
		g[i] = make([]int, len(f[i]))
	}
	return
}

func (f Fields) clone() (g Fields) {
	g = make([][]int, len(f))
	for i := range f {
		g[i] = make([]int, len(f[i]))
		for j, v := range f[i] {
			g[i][j] = v
		}
	}
	return
}

// Calculates the distance to the nearest friendly unit from the unit
// indicated by the given coordinates.  Returns 0 if no friendly units
// are reachable.
func (f Fields) distanceToNearestUnit(c Coords, player int) int {
	dist := f.cloneZero()
	queue := make([]Coords, 0, 4)
	var markUnit func(Coords, int)
	markUnit = func(c Coords, player int) {
		dist[c[0]][c[1]] = -1
		for dir := 0; dir < 4; dir++ {
			d := c.stepTo(dir)
			if d.inRange(f) && dist[d[0]][d[1]] == 0 {
				dist[d[0]][d[1]] = 1
				switch *f.get(d) {
				case player:
					markUnit(d, player)
				default:
					queue = append(queue, d)
				}
			}
		}
	}
	markUnit(c, *f.get(c))
	for qpos := 0; qpos < len(queue); qpos++ {
		c = queue[qpos]
		e := dist[c[0]][c[1]]
		switch *f.get(c) {
		case 0:
			for dir := 0; dir < 4; dir++ {
				d := c.stepTo(dir)
				if d.inRange(f) && dist[d[0]][d[1]] == 0 {
					queue = append(queue, d)
					dist[d[0]][d[1]] = e + 1
				}
			}
		case player:
			return e
		}
	}
	return 0 // no reachable friendly unit
}

func (f Fields) valid(move Move) bool {
	f = f.clone()
	player := *f.get(move[0])
	dist := f.distanceToNearestUnit(move[0], player)
	if dist == 0 {
		return false // no reachable friendly unit
	}
	const magic = 7 // unused value
	size := (&f).relabelConnected(move[0], player, magic)
	if size == 1 && move[0].distanceTo(move[1]) > 1 {
		return false // singleton must move to adjacent field
	}
	f.swap(move[0], move[1])
	return f.distanceToNearestUnit(move[1], player) < dist &&
		(&f).relabelConnected(move[1], magic, player) == size
}

func (s *State) Valid(m Move) bool {
	return m.inRange(s.Fields) && *s.Fields.get(m[0]) == s.NextPlayer() &&
		*s.Fields.get(m[1]) == 0 && s.Fields.valid(m)
}

func (s *State) Over() bool {
	// The games is over iff. the next player has no possible moves.
	ch := s.generateMaxMoves(1)
	_, ok := <-ch
	return !ok
}

func (s *State) ListMoves() (moves []interface{}) {
	for move := range s.generateAllMoves() {
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

func (fields Fields) WriteBoard(w io.Writer) (int, error) {
	var n, m int
	for _, row := range fields {
		n += len(row) + 1
	}
	buf := make([]byte, n)
	for _, row := range fields {
		for _, val := range row {
			switch val {
			case 0:
				buf[m] = '.'
			case -1:
				buf[m] = '-'
			case +1:
				buf[m] = '+'
			default:
				buf[m] = '#'
			}
			m++
		}
		buf[m] = '\n'
		m++
	}
	return w.Write(buf)
}

func (s *State) WriteLog(w io.Writer) (n int, err error) {
	for i := 0; i < len(s.History)/2 && err == nil; i++ {
		n, err = fmt.Fprintf(w, "%3d. %-8s %s\n", i+1, s.History[2*i], s.History[2*i+1])
	}
	if len(s.History)%2 == 1 && err == nil {
		n, err = fmt.Fprintf(w, "%3d. %s\n", len(s.History)/2+1, s.History[len(s.History)-1])
	}
	return
}
