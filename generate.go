package grammar

import (
	"errors"
	"fmt"
	"strings"
)

// Generates a random phrase for id based on a syntax tree.
// If id is empty the last identifier in the tree is used.
func (tree *Tree) Generate(id string) (string, error) {

	var node *node = nil
	unique := false

	// Find base node for identifier
	if len(tree.root.child) == 0 {
		return "", errors.New("empty tree")
	}

	if id == "" {
		// Empty string selects the last identifier
		node = &tree.root.child[len(tree.root.child)-1]
	} else {
		if id[0] == '*' {
			id = id[1:]
			unique = true
		}

		for i, n := range tree.root.child {
			if n.Text == id {
				node = &tree.root.child[i]
			}
		}

		if node == nil {
			return "", fmt.Errorf("no such definition: %s", id)
		}

		if len(node.child) == 0 {
			return "", fmt.Errorf("root identifier %s lacks children", id)
		}

		node = &node.child[0]
	}

	// Found a starting node, compose a phrase from it
	part, err := tree.compose(node, unique)

	if err != nil {
		return "", err
	}

	// The phrase is done, do some post-processing

	// Remove spaces before and after newlines and control tokes
	part = strings.ReplaceAll(part, " << ", "")
	part = strings.ReplaceAll(part, " <<", "")
	part = strings.ReplaceAll(part, "<< ", "")
	part = strings.ReplaceAll(part, " \n ", "\n")
	part = strings.ReplaceAll(part, " \n", "\n")
	part = strings.ReplaceAll(part, "\n ", "\n")

	// ^ capitalize the following letter, so they need to be flush
	part = strings.ReplaceAll(part, "^ ", "^")

	for p := strings.Index(part, "^"); p != -1; p = strings.Index(part, "^") {
		if p >= len(part)-1 {
			// Ignore ^ at end of string: there's nothing to uppercase
			break
		} else {
			part = part[0:p] + strings.ToUpper(string(part[p+1])) + part[p+2:]
		}
	}

	return part, nil
}

// compose builds a phrase starting from node, concatenating words
// from its children, choosing randomly among branches.
//
// If unique is true (and node is a group), picks a branch that hasn't been used before.
func (tree *Tree) compose(node *node, unique bool) (string, error) {

	if node.internalType == group {
		// Randomly pick one of the branches in the group
		opts := len(node.child)
		pick := random(0, opts-1)

		for i := 0; i < opts; i++ {
			p := &node.child[(pick+i)%opts]

			// With unique flag, keep retrying until we get something we haven't used before.
			if unique {
				if _, found := tree.uniqueUsed[p]; found {
					goto next
				}

				// This branch hasn't been used before, so it's ok.
				// Only make it as exhausted it if we are actually requesting a unique substitution!
				tree.uniqueUsed[p] = true
			}

			// Fall through by default
			return tree.compose(p, false)

		next:
		}

		// There were no unused branches remaining
		return "", errors.New("all options exhausted")
	}

	collect := []string{}

	// Only "text" nodes have their text included in the composition.
	// tag, dummy and group (already handled) don't add any text of their own.

	if node.internalType == text {
		part, err := tree.inflate(node.Text, unique)

		if err != nil {
			return "", fmt.Errorf("from %s: %s", node.Source, err)
		}

		collect = append(collect, part)
	}

	for i := range node.child {
		part, err := tree.compose(&node.child[i], false)

		if err != nil {
			return "", err
		}

		collect = append(collect, part)
	}

	ret := strings.Join(collect, " ")

	// Try to "dwim" by cleaning up spaces around punctuation
	substitutions := map[string]string{
		" )":  ")",
		"( ":  "(",
		" ,":  ",",
		" .":  ".",
		" ?":  "?",
		" !":  "!",
		" :":  ":",
		" ;":  ";",
		" _ ": " ",
		" _":  "",
		"_ ":  "",
	}

	for from, to := range substitutions {
		ret = strings.ReplaceAll(ret, from, to)
	}

	return ret, nil
}

// inflate expands the string s, substituting aliases from a syntax tree, evaluating numerical expressions, etc.
func (tree *Tree) inflate(s string, unique bool) (string, error) {

	// Scan s for a {...} sequence. This can be either;
	//
	// - a string substitution (recurse and use another key from the tree)
	// - a random number range
	//
	// Keep doing this until there are no more substitutions
	// remaining, i.e. changed remains false through the loop.

	changed := true

	for changed {
		changed = false
		resumeAt := 0
		sequenceOpen := -1

		// Scan for a {, mark its position to resume at once the substitution is handled.
		// It's very likely the substituted text contains nested substitutions so we must
		// proceed carefully, slowly bulldozing them out of the way.

		for p := resumeAt; p < len(s); p++ {
			if s[p] == '{' {
				// Found opening {; keep scanning until we get a }
				sequenceOpen = p
				resumeAt = p
			} else if s[p] == '}' {
				// Make sure the } is paired with an opening {!

				// A stray } is actually an error, but it should have been detected during parsing.
				// There's no meaningful way to report it at this stage, since we have discarded source metadata!
				if sequenceOpen >= 0 {
					replace := s[sequenceOpen : p+1]

					var replaceWith string = "(ERROR)"
					var err error
					var bottomBound, topBound int

					if replace == "{\\n}" {
						replaceWith = "\n"
					} else if _, err = fmt.Sscanf(replace, "{%d-%d}", &bottomBound, &topBound); err == nil {
						replaceWith = fmt.Sprintf("%d", random(bottomBound, topBound))
					} else {
						tag := s[sequenceOpen+1 : p]

						replaceWith, err = tree.Generate(tag)

						if err != nil {
							return "", fmt.Errorf("%s (%s)", err, tag)
						}
					}

					//s = strings.Replace(s, replace, replaceWith, 1)
					s = s[0:sequenceOpen] + replaceWith + s[p+1:]
					changed = true
					break
				}
			}
		}
	}

	return s, nil
}
