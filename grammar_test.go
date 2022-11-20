package grammar

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

// Check that parsing works. These are mostly the examples used in the docs.
func TestParsing(t *testing.T) {
	input := []string{
		`abc
                 [
                   [a|b] c
                   | [a|b] c
                   | bcd
                 ]`,
		`a [ b ]
                 abc
                 [
                   {a}
                   | ccc [ {a} | {a} ]
                   | {a}
                   | [ aaa ]
                   | {a} {a}
                 ]
                `,
		`abc [ [aaa | bbb] | ccc ]`,
		`identifier [ text ]`,
		`greeting [ hi | hello there | good morning to thee ]`,
		`diary
                 [
                   It was [Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday],
                   the [first|second|third|fourth] week of
                   [January|February|March|April|May|June|July|August|September|October|November|December].
                 ]`,
		`weekday [ Monday | Tuesday | Wednesday | Thursday | Friday | Saturday | Sunday ]
                 month   [ January | February | March | April | May | June | July | August | September | October | November | December ]
                 ordinal [ first|second|third|fourth ]
                 diary   [ It was {weekday} , the {ordinal} week of {month}. I had just had my {ordinal} cup of coffee for the day... ]`,
		`lines   [ This is a line. {\n} This is another line.]`,
		`verdict [ I'm not angry, but I'm [very|_] disappointed. ]`,
		`weekday [ [Mon|Tues|Wednes|Thurs|Fri|Satur|Sun] << day, next week? ]  // "Tuesday, next week?", not "Tues day, next week?"`,
		`excuse  [ My [dog | cat] ate my homework. ]  // What a jerk!!`,
		`long_month        [ {1-31} ]
                 short_month       [ {1-30} ]
                 very_short_month  [ {1-28} ]
                 date
                 [
                   [
                     Jan {long_month}  | Feb {very_short_month} | Mar {long_month}  |
                     Apr {short_month} | May {long_month}       | Jun {short_month} |
                     Jul {long_month}  | Aug {long_month}       | Sep {short_month} |
                     Oct {long_month}  | Nov {short_month}      | Dec {long_month}
                   ], {1980-2020}
                 ]`,
		`measure     [ dl | tbsp | tsp ]
                 ingredient  [ fluor | sugar | salt | yeast ]
                 recipe
                 [
                   {1-6} {measure} {*ingredient} {\n}
                   {1-6} {measure} {*ingredient} {\n}
                   {1-6} {measure} {*ingredient} {\n}
                   {1-6} {measure} {*ingredient}
                 ]
                `,
		`recursive [ (ran!) {recursive} | stop ]`,
		`phonetic  [ Alpha | Bravo | Foxtrot | Quebec | Whiskey | Xray ]
                 code      [ {*phonetic} {phonetic} {phonetic} {\n} ]
                 message   [ {code} {code} {code} {code} {code} {code} ]
                `,
		`promise [ I'll do it first thing tomorrow ([j/k, lol | maybe | for [real|sure] !]) ]`,
		"\ta[b]",
		" \ta[b]",
		"\t a[b]",
		"\t\ta[b]",
		" \t a[b]",
		`	a[b] // literal tab at start of line`,
		"a[b]//ignore",
		"a[[[[[[[[[[[[[[[[[[[[b]]]]]]]]]]]]]]]]]]]]",
		"where [ ^ here and ^ there ]",
		"abc [ abc^ ]",
		"headline [ {5-25} [Cute | Adorable | Inspiring | Weird] Photos Of [Cats | Celebrities | Two-Stroke Tractors] You Haven't Seen Before ]",
	}

	for _, in := range input {
		var err error
		var tree *Tree
		var out string

		tree, err = Parse(in)

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		t.Logf("Tree:\n%s\n", tree.Format())

		out, err = tree.Generate("")

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		t.Logf("Output:\n%s\n", out)
	}
}

// Test that Parse() creates a trees with the correct number of nodes
func TestCountNodes(t *testing.T) {

	input := map[string]int{
		"a[[b]]":   5,
		"a[[b]c]":  6,
		"a[b|c|d]": 5,
	}

	for in, expectedCount := range input {
		tree, err := Parse(in)

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		if count := tree.Count(); count != expectedCount {
			t.Logf("\n%s", tree.Format())
			t.Fatalf("\"%s\" has wrong Node.Count (expected %d, got %d)", in, expectedCount, count)
		}
	}
}

// Provoke syntax errors
func TestParsingErrors(t *testing.T) {

	// None of these should work
	badInput := []string{
		"[a|b]",
		"a[]",
		"]",
		"a[a]]",
		"a[",
		"a[a|b",
		"a[a|",
		"a[|a]",
		"a[a||a]",
		"|",
		"a|[a]",
		"a[a]|",
		"{a[a]",
		"a}[a]",
		"<a[b]",
		"a<[b]",
		"a b[c]",
		"a[{b]",
		"a[{a b}]",
		"a[{b",
		"a {b",
		"a[{b|c]",
		"a[b}]",
		"//a[b]",
		"*a[b]",
		"a*[b]",
		"a*b[c]",
		"a[b] a[c]",
		"a[b}]",
		"a[b",
	}

	for _, in := range badInput {
		_, err := Parse(in)

		t.Logf("%-10s => %s\n", in, err)

		if err == nil {
			t.Fatalf("\"%s\" should have failed, but didn't", in)
		}
	}
}

// Test that input matches an expected set of output
func TestGenerate(t *testing.T) {

	// Value is a slice of valid outputs; can be a single one
	input := map[string][]string{
		"a[b]":         {"b"},
		"a[b|c]":       {"b", "c"},
		"a[b|c|d]":     {"b", "c", "d"},
		"a[[b|c] d]":   {"b d", "c d"},
		"a[[b|c]<<d]":  {"bd", "cd"},
		"a[[b].]":      {"b."},
		"a[[b],]":      {"b,"},
		"a[[b]:]":      {"b:"},
		"a[[b];]":      {"b;"},
		"a[[b]!]":      {"b!"},
		"a[[b] << c]":  {"bc"},
		"a[[b]<< c]":   {"bc"},
		"a[[b] <<c]":   {"bc"},
		"a[b{\\n}c]":   {"b\nc"},
		"a[< << <]":    {"<<"},
		"a[/ << /]":    {"//"},
		"a[ ( b ) ]":   {"(b)"},
		"a[^b]":        {"B"},
		"c[b] a[^{c}]": {"B"},
	}

	for in, validOutput := range input {
		var err error
		var tree *Tree
		var out string

		tree, err = Parse(in)

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		out, err = tree.Generate("")

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		for _, v := range validOutput {
			if out == v {
				goto next
			}
		}

		t.Fatalf("\"%s\" failed (expected \"%s\", got \"%s\")", in, validOutput, out)

	next:
	}
}

// Miscellaneous error checks
func TestGenerateErrors(t *testing.T) {
	brokenTree := Tree{root: node{}}

	// t.Logf("Passing tree with %d children to Generate()", len(brokenRoot.Child));

	if _, err := brokenTree.Generate(""); err == nil {
		t.Fatalf("Generate() should have failed (empty tree), but didn't")
	}

	workingTree, _ := Parse("a[a]")

	if _, err := workingTree.Generate("missing"); err == nil {
		t.Fatalf("Generate() should have failed (missing id), but didn't")
	}
}

// Make sure substitutions prefixed with * only give the same result once
func TestUniqueSubst(t *testing.T) {

	in := "a[b|c|d] e[{*a}{*a}{*a}] f[{*a}] g[{*a}{*a}{*a}{*a}]"

	var tree *Tree
	var err error
	var out string

	// This should be enough...
	for i := 0; i < 100; i++ {
		tree, err = Parse(in)

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		out, err = tree.Generate("e")

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		t.Logf("%s => %s\n", in, out)

		if strings.Count(out, "b") != 1 || strings.Count(out, "c") != 1 || strings.Count(out, "d") != 1 {
			t.Fatalf("unique substitution failed: \"%s\" => \"%s\"", in, out)
		}

		// This should fail, since all options have been exhausted already
		exhaustedID := "f"
		_, err = tree.Generate(exhaustedID)

		if err == nil {
			t.Fatalf("\"%s\" in \"%s\" should have been exhausted already", exhaustedID, in)
		} else {
			t.Logf("Got (expected!) error: %s", err)
		}

		tree.Reset()

		// This should also fail
		exhaustedID = "g"
		_, err = tree.Generate(exhaustedID)

		if err == nil {
			t.Fatalf("\"%s\" in \"%s\" should have been exhausted already", exhaustedID, in)
		} else {
			t.Logf("Got (expected!) error: %s", err)
		}
	}
}

func TestTreeFormatOptions(t *testing.T) {
	tree, _ := Parse("a[b]")

	if strings.Count(tree.Format(DisplayGroupNumbers), "[1") != 1 {
		t.Error("Format() didn't include branchnumber")
	}

	if strings.Count(tree.Format(), "[1") != 0 {
		t.Error("Format() had branchnumber, even though it shouldn't")
	}

	if strings.Count(tree.Format(DisplaySource), ":1") == 0 {
		t.Error("Format() didn't include source")
	}

	if strings.Count(tree.Format(), ":1") != 0 {
		t.Error("Format() had source, even though it shouldn't")
	}
}

// Make sure Generate() called with *identifier returns the same output only once
func TestGenerateExclusive(t *testing.T) {

	in := "a [b | c | d | e | f]"

	var tree *Tree
	var err error
	var out string

	tree, err = Parse(in)

	if err != nil {
		t.Fatalf("\"%s\" failed (%s)", in, err)
	}

	previous := ""

	for i := 0; i < 5; i++ {
		out, err = tree.Generate("*a")

		if err != nil {
			t.Fatalf("\"%s\" failed (%s)", in, err)
		}

		t.Logf("%s => \"%s\", \"%s\"", in, out, previous)

		if strings.Contains(previous, out) {
			t.Fatalf("\"%s\" has already been generated (\"%s\")", out, previous)
		}

		previous += out
	}
}
