package grammar

import (
	"fmt"
)

type nodeType int

const (
	unknown nodeType = iota
	root
	text
	group
	dummy
	tag
)

type node struct {
	internalType nodeType
	Text         string
	child        []node
	Source       string // Where this token originated
}

// Returns a text representation of an individual node.
//
// Note that this is different from Format, which formats a whole tree.
func (node *node) formatNode(options []TreeFormatOption) string {
	switch node.internalType {
	case root:
		return "(root)"
	case text:
		return node.Text
	case tag:
		return node.Text
	case group:
		if hasOption(DisplayGroupNumbers, options) {
			return node.Text
		} else {
			return "["
		}
	case dummy:
		return "*"
	default:
		return "?"
	}
}

type TreeFormatOption int

const (
	// Include source file and line number for each token
	DisplaySource TreeFormatOption = iota
	// Include unique group IDs (e.g. [23)
	DisplayGroupNumbers
)

func hasOption(find TreeFormatOption, in []TreeFormatOption) bool {
	for _, option := range in {
		if option == find {
			return true
		}
	}

	return false
}

// add adds definitions to a grammar syntax tree.
func (root *node) add(path []string, source string, nodeType nodeType) (*node, error) {
	group := root

	for {
		// If this is the last element in the stack, add it last on the current group
		if len(path) == 1 {
			add := node{Text: path[0], Source: source, internalType: nodeType}
			group.child = append(group.child, add)
			return &add, nil
		}

		// Otherwise, search the tree for the next element in the path
		find := path[0]

		for i := len(group.child) - 1; i >= 0; i-- {
			node := &group.child[i]

			if node.Text == find {
				group = node
				// Drop the left-most element of the search path
				path = path[1:]
				goto next
			}
		}

		return nil, fmt.Errorf("incorrect path: missing %s under %s", find, group.Text)

	next:
	}
}
