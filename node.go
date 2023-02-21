package latex

type Kind int

const (
	TextKind = iota
	DocumentKind
	ElementKind
)

type Node struct {
	Kind       Kind
	Parameters map[string]string
	Data       string
	Children   []*Node
}
