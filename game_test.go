package ayu

import "testing"
import "bytes"

func testParseCoords(t *testing.T, input string, expect_c Coords, expect_ok bool) {
	c, ok := ParseCoords(input)
	if c != expect_c || ok != expect_ok {
		t.Error(input, expect_c, c, expect_ok, ok)
	}
}

var goodCoords = map[string]Coords{
	"A1":  Coords{10, 0},
	"K11": Coords{0, 10},
	"A11": Coords{0, 0},
	"K1":  Coords{10, 10},
	"E2":  Coords{9, 4}}

var badCoords = []string{
	"A0", "A12", "@1", "L1", "A1\n", "e2", "", "xyzzy"}

func TestParseCoords(t *testing.T) {
	for input, c := range goodCoords {
		testParseCoords(t, input, c, true)
	}
	for _, input := range badCoords {
		testParseCoords(t, input, Coords{}, false)
	}
}

func testParseMove(t *testing.T, input string, expect_m Move, expect_ok bool) {
	m, ok := ParseMove(input)
	if m != expect_m || ok != expect_ok {
		t.Error(input, expect_m, m, expect_ok, ok)
	}
}

func TestParseMove(t *testing.T) {
	for i, src := range goodCoords {
		for j, dst := range goodCoords {
			testParseMove(t, i+"-"+j, Move{src, dst}, true)
		}
	}
	for i, _ := range goodCoords {
		for _, j := range badCoords {
			testParseMove(t, i+"-"+j, Move{}, false)
			testParseMove(t, j+"-"+i, Move{}, false)
		}
	}
	for _, i := range badCoords {
		for _, j := range badCoords {
			testParseMove(t, i+"-"+j, Move{}, false)
		}
	}
	testParseMove(t, "A7-C3", Move{Coords{4, 0}, Coords{8, 2}}, true)
	testParseMove(t, "A7C3", Move{}, false)
	testParseMove(t, "A7,C3", Move{}, false)
	testParseMove(t, "A7~C3", Move{}, false)
}

func TestCreateState(t *testing.T) {
	state := CreateState()
	var b bytes.Buffer
	state.WriteBoard(&b)
	if string(b.Bytes()) != `.+.+.+.+.+.
-.-.-.-.-.-
.+.+.+.+.+.
-.-.-.-.-.-
.+.+.+.+.+.
-.-.-.-.-.-
.+.+.+.+.+.
-.-.-.-.-.-
.+.+.+.+.+.
-.-.-.-.-.-
.+.+.+.+.+.
` {
		t.Error(string(b.Bytes()))
	}
}
