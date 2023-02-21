package latex

import "errors"

// stringify extracts text from array of nodes or returns error if there are non-text nodes
func stringify(children []*Node) (str string, err error) {
	for _, child := range children {
		if child.Kind != TextKind {
			return "", errors.New("only text is allowed here")
		}

		str += child.Data
	}

	return
}

func isNewline(name string) bool {
	return name == "\\\\" || name == "\\newline" || name == "\\*"
}
