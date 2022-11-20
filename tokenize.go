package grammar

import (
	"strings"
	"fmt"
)

type token struct {
	Text   string
	Source string
}

// tokenize splits an input grammar string and returns a slice of Token containing the individual words. Syntactic
// characters [ | ] are separated from surrounding text. Each Token is also flagged with its source file (as provided by
// the file argument) and line number to facilitate error handling. No syntactical meaning is assigned to the tokens at
// this time; only the raw text is returned.
func tokenize(input string, file string) []token {
	var ret []token

	for lineNo, line := range strings.Split(input, "\n") {
		// Process input line by line

		var collect []token
		source := fmt.Sprintf("%s:%d", file, lineNo+1) // Physical line number

		// Strip whitespace
		line = strings.ReplaceAll(line, "\t", "")

		line = strings.Trim(line, " ")

		// Add extra spaces around syntactic characters so they will separated properly
		line = strings.Replace(line, "//", " // ", -1)
		line = strings.Replace(line, "[", " [ ", -1)
		line = strings.Replace(line, "]", " ] ", -1)
		line = strings.Replace(line, "|", " | ", -1)
		line = strings.Replace(line, "{", " {", -1)
		line = strings.Replace(line, "}", "} ", -1)
		line = strings.Replace(line, "  ", " ", -1)

		for _, t := range strings.Split(line, " ") {
			t = strings.Trim(t, " ")

			if t == "//" {
				// Discard the rest of the line, but save what we already collected
				ret = append(ret, collect...)
				goto next_line
			} else if t != "" {
				collect = append(collect, token{Text: t, Source: source})
			}
		}

		ret = append(ret, collect...)
	next_line:
	}

	return ret
}
