package ayu

import "bytes"
import "strings"
import "testing"

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
	(&state.Fields).WriteBoard(&b)
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

func testWriteLog(t *testing.T, moves string, expected string) {
	state := CreateState()
	for _,part := range(strings.Fields(moves)) {
		if move, ok := ParseMove(part); !ok {
			t.Error("Could not parse move:", part)
		} else if !state.Execute(move) {
			t.Error("Could not execute move:", move)
		}
	}
	var b bytes.Buffer
	state.WriteLog(&b)
	if b.String() != expected {
		t.Error(`"` + expected  + `"`, `"` + b.String() + `"`)
	}
}

func TestWriteLog(t *testing.T) {
	testWriteLog(t, "", "")
	testWriteLog(t, "D9-E9",
		`  1. D9-E9
`)
	testWriteLog(t, "D9-E9 E10-F10",
		`  1. D9-E9    E10-F10
`)
	testWriteLog(t, "D9-E9 E10-F10 B9-B10",
		`  1. D9-E9    E10-F10
  2. B9-B10
`)
	testWriteLog(t, "D9-E9 E10-F10 B9-B10 A6-A7",
		`  1. D9-E9    E10-F10
  2. B9-B10   A6-A7
`)
	testWriteLog(t, "D9-E9 E10-F10 B9-B10 A6-A7 J11-J10 C10-C9",
		`  1. D9-E9    E10-F10
  2. B9-B10   A6-A7
  3. J11-J10  C10-C9
`)
}
