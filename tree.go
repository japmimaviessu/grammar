package grammar

import (
	"fmt"
	"strings"
)

// A Tree represents a grammar syntax tree.
type Tree struct {
	root       node
	uniqueUsed map[(*node)]bool
}

// Count returns the number of nodes in a syntax tree.
func (tree *Tree) Count() int {
	return tree.root.count()
}

func (node *node) count() int {
	count := 0

	for _, child := range node.child {
		count += 1 + child.count()
	}

	return count
}

// Format returns a string visualizing a grammar tree.
//
// Accepts any number of [TreeFormatOption] to alter the output.
func (tree *Tree) Format(options ...TreeFormatOption) string {
	rawLines := tree.root.internalFormat("", options)
	lines := treeLines(rawLines, options)
	return strings.Join(lines, "\n")
}

// For formatting a line with a left/right column
type formatLine struct {
	left  string
	right string
}

// internalFormat recursively indents node with spaces and box-drawing characters.
func (node *node) internalFormat(prefix string, options []TreeFormatOption) []formatLine {
	var collect []formatLine

	for _, node := range node.child {
		// Describe this node. Put source in the right column; decide later if we'll use it.
		collect = append(collect, formatLine{prefix + "└─ " + node.formatNode(options), node.Source})
		// Ask children to describe themselves. Nudge them a bit to the right by adding to the prefix.
		collect = append(collect, node.internalFormat(prefix + "   ", options)...)
	}

	return collect
}

// treeLines beautifies a syntax tree with box-drawing characters.
func treeLines(input []formatLine, options []TreeFormatOption) []string {
	lines := len(input)
	runes := make([][]rune, lines)

	maxWidth := 0

	// Convert each line to runes and check for the widest line
	for i := 0; i < lines; i++ {
		runes[i] = []rune(input[i].left)

		if len(runes[i]) > maxWidth {
			maxWidth = len(runes[i])
		}
	}

	connected := make([]bool, maxWidth)

	// Scan lines bottom-up, then columns left-right.
	// A bottom-left corner will mark the column as connected.
	// A bottom-left corner in an already connected column is changed into a left-tee
	// A space in a connected row is connected to a vertical line
	// If we run into anything that isn't a bottom-left corner or
	// space, it's a leaf with text. Reset the connected flag for the column.
	for i := lines - 1; i >= 0; i-- {
		rl := &runes[i]

		thisLen := len(*rl)

		for j := 0; j < maxWidth; j++ {
			if j >= thisLen {
				connected[j] = false
				continue
			}

			r := &(*rl)[j]

			switch {
			case *r != '└' && *r != ' ':
				connected[j] = false
			case *r == '└' && connected[j]:
				*r = '├'
			case *r == ' ' && connected[j]:
				*r = '│'
			case *r == '└':
				connected[j] = true
			}
		}
	}

	// Convert runes back to strings. Pad & append source lines, if requested.
	ret := make([]string, lines)

	if hasOption(DisplaySource, options) {
		for i := 0; i < lines; i++ {
			ret[i] = fmt.Sprintf("%-*s%s", maxWidth, string(runes[i][3:]), input[i].right)
		}
	} else {
		for i := 0; i < lines; i++ {
			ret[i] = string(runes[i][3:])
		}
	}

	return ret
}

// Reset clears the list of used unique substitutions.
func (tree *Tree) Reset() {
	tree.uniqueUsed = make(map[*node]bool)
}
