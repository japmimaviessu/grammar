// Package grammar provides a generative grammar parser for composing random phrases.
//
// # Basic Usage
//
//   - Parse() a grammar, resulting in a syntax tree
//   - Generate() a random phrase (for a given identifier) based on the syntax tree
//
// For convenience, there is a shortcut Quick() that both parses and generates, but ignores any errors.
//
// # Input Format
//
// A grammar contains one or more phrase definitions, comprised of an identifier followed by some text.
//
//	identifier [ text ]
//
// Words are concatenated sequentially when the grammar is evaluated. Each [ ] clause forms a group, which may contain
// one or more branches, delimited by |. These are chosen among at random during evaluation. Groups may contain nested
// groups.
//
// Generate() composes a random phrase by traversing the syntax tree, randomly selecting a path among the branches and
// stiching together the words encountered.
//
//	greeting [ hello there | good [morning | evening] ]
//
// This will output either of the three phrases "hello there", "good morning" or "good evening".
//
// Whitespace is ignored and definitions may span any number of lines:
//
//	diary
//	[
//	  It was [Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday],
//	  the [first|second|third|fourth] week of
//	  [January|February|March|April|May|June|July|August|September|October|November|December].
//	]
//
// Sample output:
//
//	It was Friday, the first week of March.
//
// To give a better understanding what is going on internally, Format() can visualize the syntax tree:
//
//	diary
//	└─ [
//	   └─ It was
//	      ├─ [
//	      │  ├─ Monday
//	      │  ├─ Tuesday
//	      │  ├─ Wednesday
//	      │  ├─ Thursday
//	      │  ├─ Friday
//	      │  ├─ Saturday
//	      │  └─ Sunday
//	      └─ , the
//	         ├─ [
//	         │  ├─ first
//	         │  ├─ second
//	         │  ├─ third
//	         │  └─ fourth
//	         └─ week of
//	            ├─ [
//	            │  ├─ January
//	            │  ├─ February
//	            │  ├─ March
//	            │  ├─ April
//	            │  ├─ May
//	            │  ├─ June
//	            │  ├─ July
//	            │  ├─ August
//	            │  ├─ September
//	            │  ├─ October
//	            │  ├─ November
//	            │  └─ December
//	            └─ .
//
// To make the grammar more readable we can separate groups into multiple definitions and perform substitutions by
// wrapping an identifier with a { } substitution marker. This will make a detour to evaluate that identifier and insert
// the result in place of the marker.
//
//	weekday [ Monday | Tuesday | Wednesday | Thursday | Friday | Saturday | Sunday ]
//	month   [ January | February | March | April | May | June | July | August | September | October | November | December ]
//	ordinal [ first | second | third | fourth ]
//	diary   [ It was {weekday}, the {ordinal} week of {month}.  ]
//
// This also allows reusing the same definition multiple times:
//
//	diary   [ It was {weekday}, the {ordinal} week of {month}. I had just had my {ordinal} cup of coffee for the day... ]
//
// // can be used for comments; anything to the end of line is ignored.
//
//	excuse  [ My [dog | cat] ate my homework. ]  // What a jerk!!
//
// # Special Formatting
//
// While sentence structure and punctuation can appear somewhat butchered in the syntax tree visualization, Generate()
// tries to do what is reasonable while stitching it all together. Output is concatenated word-by-word with single
// spaces in between. Punctuation (.,:;!?) is left-adjusted to the preceding word and parentheses ( ) are tightened to
// the nearest word.
//
//	promise [ I'll do it first thing tomorrow ([j/k, lol | maybe | for [real|sure] !]) ]
//
// Sample output:
//
//	I'll do it first thing tomorrow (for real!)
//
// To remove unwanted spaces, there is the "force concatenation" token <<:
//
//	weekday [ [Mon|Tues|Wednes|Thurs|Fri|Satur|Sun] << day, next week? ]  // "Tuesday, next week?", not "Tues day, next week?"
//
// Newlines can be explicitly inserted with {\n}. Note that spaces are omitted before and after {\n}, akin to <<.
//
//	lines [ This is a line. {\n} This is another line. ]  // "This is a line.\nThis is another line."
//
// An empty group or branch is a syntax error. The special "empty" token _ can be used to explicitly omit output:
//
//	verdict [ I'm not angry, but I'm [very | _] disappointed. ]
//
// ^ will convert the following character to uppercase:
//
//	where [ ^ here and ^ there ]  // Here and There
//
// # Substitution Options
//
// Substitution can generate random numbers by specifying an interval:
//
//      headline [ {5-25} [Cute | Adorable | Inspiring | Weird] Photos Of [Cats | Celebrities | Two-Stroke Tractors] You Haven't Seen Before ]
//
// Naturally, substitutions can be nested:
//
//      long_month        [ {1-31} ]
//      short_month       [ {1-30} ]
//      very_short_month  [ {1-28} ]
//
//      date
//      [
//        [
//          Jan {long_month}  | Feb {very_short_month} | Mar {long_month}  |
//          Apr {short_month} | May {long_month}       | Jun {short_month} |
//          Jul {long_month}  | Aug {long_month}       | Sep {short_month} |
//          Oct {long_month}  | Nov {short_month}      | Dec {long_month}
//        ], {1980-2020}
//      ]
//
// Sample output:
//
//      Aug 29, 1996
//
// You can mark a substitution as exclusive by prefixing the identifier with *:
//
//	measure     [ dl | tbsp | tsp ]
//	ingredient  [ fluor | sugar | salt | yeast ]
//	recipe
//	[
//	  {1-6} {measure} {*ingredient} {\n}
//	  {1-6} {measure} {*ingredient} {\n}
//	  {1-6} {measure} {*ingredient} {\n}
//	  {1-6} {measure} {*ingredient}
//	]
//
// This will ensure each top-level branch of the ingredient identifier is used only once:
//
//	1 tbsp fluor
//	3 dl salt
//	3 tsp sugar
//	6 tbsp yeast
//
// Exclusive substitutions will fail with an error if an identifier is requested more times than there are branches in
// its top-level group. Note that exclusive substitution are only enforced for identifiers prefixed with *. These can be
// mixed with regular (non-exclusive) substitions, which don't care if the identifier has been used before:
//
//	phonetic  [ Alpha | Bravo | Foxtrot | Quebec | Whiskey | Xray ]
//	code      [ {*phonetic} {phonetic} {phonetic} {\n} ]
//	message   [ {code} {code} {code} {code} {code} {code} ]
//
// Sample output:
//
//	Xray Bravo Foxtrot
//	Foxtrot Quebec Alpha
//	Bravo Alpha Foxtrot
//	Whiskey Whiskey Bravo
//	Alpha Xray Bravo
//	Quebec Alpha Bravo
//
// The exclusive substitution list will persist between calls to Generate(). It can be cleared with Reset(). The *
// prefix can also be used directly in calls to Generate().
//
package grammar

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

// Parse parses an input grammar string and returns a syntax tree.
//
// If a syntax error is encountered it returns an error and an empty string.
func Parse(grammar string) (*Tree, error) {
	return parseInternal(tokenize(grammar, ""))
}

// ParseFile reads and parses an input grammar from filename and returns a syntax tree.
func ParseFile(filename string) (*Tree, error) {
	return ParseFiles([]string{filename})
}

// ParseFiles reads and parses an input grammar from multiple files and returns a syntax tree. Files are processed
// individually, not concatenated, so each file must be self-contained and syntactically complete. Note that if any of
// the files contains an error the whole operation will fail.
func ParseFiles(filenames []string) (*Tree, error) {
	var token []token

	for _, f := range filenames {
		contents, err := ioutil.ReadFile(f)

		if err != nil {
			return nil, err
		}

		moreTokens := tokenize(string(contents), f)

		if err != nil {
			return nil, err
		}

		token = append(token, moreTokens...)
	}

	return parseInternal(token)
}

// parseInternal parses an input grammar in the form of a slice of input tokens and constructs a syntax tree.
//
// Dummy nodes are sometimes required to represent nested groups. Where a group opens with another group, followed by
// text - e.g. [[a|b]c] - a dummy node is inserted between [ and [ to provide an anchor point for c. There should only
// ever need to be one dummy node per group. For simplicity, dummy nodes are also added even where they are superfluous,
// e.g. [[a|b]]. Dummy nodes internally have the text // which is used for comments and should never be found in the
// tree otherwise.
//
// Since there are often multiple sequential group, group nodes are assigned a unique identifier ([ + number) to enable
// unambiguous paths. In the formatted print, these numbers are suppressed unless the IncludeGroupNumbers option is set.
func parseInternal(token []token) (*Tree, error) {
	if len(token) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	var root node = node{Text: "", internalType: root}
	groupID := 0        // unique ID; incremented when used
	stack := []string{} // used to keep track of the current tree path
	collect := ""
	previousSource := "" // syntax errors are sometimes at the previous token, not the current

	// Iterate over input tokens. Scan for [ | ] control tokens; everything else is concatenated onto collect. When
	// a control token is encountered there should be *something* in collect or it is a syntax error.

	// After any control token collect should be set to empty.

	for _, t := range token {

		// These should have been removed by tokenize()!
		if t.Text == "" {
			return nil, errors.New("empty token")
		}

		source := t.Source

		//fmt.Println(stack, ">", t.Text);

		if t.Text == "[" {
			if collect == "" && len(stack) == 0 {
				return nil, fmt.Errorf("missing definition identifier at %s", t.Source)
			} else if collect == "" && len(stack) > 1 && stack[len(stack)-1][0] == '[' {
				// [ after [ without anything in between - need to insert a dummy node
				stack = append(stack, "//")
				root.add(stack, source, dummy)
			} else if collect != "" {
				if len(stack) == 0 {
					for _, s := range root.child {
						if s.Text == collect {
							return nil, fmt.Errorf("duplicate identifier \"%s\" at %s and %s",
								t.Text, s.Source, t.Source)
						}
					}
				}

				stack = append(stack, collect)
				collect = ""

				// Top-level nodes get the "tag" type; these are purely labels
				// and its text won't be included by compose()!
				if len(stack) == 1 {
					root.add(stack, previousSource, tag)
				} else {
					root.add(stack, previousSource, text)
				}
			}

			stack = append(stack, fmt.Sprintf("[%d", next(&groupID)))
			root.add(stack, source, group)
		} else if t.Text == "|" {
			if len(stack) == 0 {
				return nil, fmt.Errorf("stray | at root level at %s", t.Source)
			} else if collect == "" && len(stack) > 0 && stack[len(stack)-1][0] == '[' {
				// If there has been nothing collected since the last
				// control token, AND we are currently in a group
				return nil, fmt.Errorf("stray | in group at %s", t.Source)
			}

			if stack[len(stack)-1][0] != '[' && collect != "" {
				root.add(append(stack, collect), source, text)
				collect = ""
			}

			// Unwind to the most recent group
			for i := len(stack) - 1; i >= 0; i-- {
				s := stack[i]

				if s[0] == '[' {
					break
				}

				stack = stack[:(len(stack) - 1)]
			}

			if collect == "" && stack[len(stack)-1][0] != '[' {
				return nil, fmt.Errorf("stray | in group at %s", t.Source)
			} else if collect != "" {
				// Add the current stack + the token(s) collected since
				// the last control character, to add it under the current group
				root.add(append(stack, collect), source, text)
				collect = ""
			}

			// [ ] directly followed by |; do not add an empty text token

		} else if t.Text == "]" {
			if collect == "" && len(stack) == 0 {
				return nil, fmt.Errorf("stray ] at %s", t.Source)
			} else if collect == "" && len(stack) > 0 && stack[len(stack)-1][0] == '[' {
				return nil, fmt.Errorf("empty group at %s", t.Source)
			} else if collect != "" {
				root.add(append(stack, collect), previousSource, text)
				collect = ""
			}

			// Scan the stack top-down, pop anything that isn't a group open [
			// and stop after the first group open we encounter
			for i := len(stack) - 1; i >= 0; i-- {
				s := stack[i]

				stack = stack[:(len(stack) - 1)]

				if s[0] == '[' {
					break
				}
			}

			// If we are back at the top-level identifier, wipe the stack
			if len(stack) == 1 {
				stack = []string{}
			}
		} else {
			if collect == "" {
				if len(stack) == 0 {
					// Use separate strings and Contains rather than ContainsAny,
					// since we want to know specifically which character was encountered
					invalidInIdentifier := []string{"{", "}", "<", "*", "^"}

					for _, find := range invalidInIdentifier {
						if strings.Contains(t.Text, find) {
							return nil, fmt.Errorf("invalid character %s in identifier at %s", find, t.Source)
						}
					}
				}

				collect = t.Text
			} else if len(stack) == 0 {
				return nil, fmt.Errorf("expecting [ after identifier at %s", t.Source)
			} else {
				collect += " " + t.Text
			}

			if t.Text[0] == '{' && t.Text[len(t.Text)-1] != '}' {
				return nil, fmt.Errorf("unterminated substitution \"%s\" at %s", t.Text, t.Source)
			} else if t.Text[0] != '{' && t.Text[len(t.Text)-1] == '}' {
				return nil, fmt.Errorf("stray } (substitution missing { ?) at %s", t.Source)
			}
		}

		previousSource = source
	}

	// We're out of tokens; make sure the last group was closed properly
	if len(stack) > 0 {
		return nil, fmt.Errorf("unterminated [ at %s", previousSource)
	}

	tree := Tree{root: root}
	tree.Reset()

	return &tree, nil
}

// Quick parses a grammar and generates the default (last) definition.
//
// Note: this will discard any errors encountered.
func Quick(grammar string) string {
	tree, err := Parse(grammar)

	if err != nil {
		return ""
	}

	ret, _ := tree.Generate("")
	return ret
}
