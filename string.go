package latex

func String(node *Node) (out string) {
	if node.Kind == TextKind {
		return node.Data
	}

	for _, child := range node.Children {
		out += String(child)
	}

	return
}
