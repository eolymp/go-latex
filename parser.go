package latex

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var identifier = regexp.MustCompile("^\\\\[a-zA-Z]+$")

type Parser struct {
	tokens *Tokenizer
	defs   map[string]string
}

func Parse(r io.RuneScanner) (*Node, error) {
	return NewParser(r).Parse()
}

func NewParser(r io.RuneScanner) *Parser {
	return &Parser{tokens: NewTokenizer(r), defs: map[string]string{}}
}

func (p *Parser) Define(key, val string) {
	p.defs[key] = val
}

func (p *Parser) Value(key string) string {
	return p.defs[key]
}

func (p *Parser) Parse() (*Node, error) {
	children, _, err := p.vertical(func(a any, err error) bool {
		return err == io.EOF
	})

	if err != nil {
		return nil, err
	}

	return &Node{Kind: DocumentKind, Children: children}, nil
}

// horizontal collects text span nodes, it expects to discover text fragments which will be displayed horizontally (one next to another)
func (p *Parser) horizontal(stop func(any, error) bool) (children []*Node, err error) {
	for {
		t, err := p.tokens.Token()
		if stop(t, err) {
			return children, nil
		}

		if err != nil {
			return nil, err
		}

		node, inline, err := p.parse(t)
		if err != nil {
			return nil, err
		}

		if node == nil {
			continue
		}

		if !inline {
			return nil, errors.New("block token in horizontal mode")
		}

		// merge consequent text nodes together
		if node.Kind == TextKind && len(children) > 0 && children[len(children)-1].Kind == TextKind {
			children[len(children)-1].Data += node.Data
			continue
		}

		children = append(children, node)
	}
}

// vertical stacks block nodes, it expects to discover paragraphs and blocks which will be displayed vertically (one below another)
func (p *Parser) vertical(stop func(any, error) bool) (children []*Node, last any, err error) {
	floating := &Node{Kind: ElementKind, Data: "\\par"}
	newline := false

	flush := func() {
		if len(floating.Children) == 0 {
			return
		}

		children = append(children, floating)
		floating = &Node{Kind: ElementKind, Data: "\\par"}
	}

	// add whatever is hanging in floating paragraph before return
	defer flush()

	for {
		t, err := p.tokens.Token()
		if stop(t, err) {
			return children, t, nil
		}

		if err != nil {
			return nil, nil, err
		}

		node, inline, err := p.parse(t)
		if err != nil {
			return nil, nil, err
		}

		if node == nil {
			continue
		}

		if !inline {
			flush()
			children = append(children, node)
			continue
		}

		// flush floating paragraph
		empty := node.Kind == TextKind && strings.TrimSpace(node.Data) == "" && strings.HasSuffix(node.Data, "\n")
		par := node.Kind == ElementKind && node.Data == "\\par" && len(node.Children) == 0
		if par || (newline && empty) {
			flush()
			continue
		}

		// remember if this line ends with \n, if next one is empty line we will start new paragraph (condition above)
		newline = node.Kind == TextKind && strings.HasSuffix(node.Data, "\n")

		// merge consequent text nodes together
		if node.Kind == TextKind && len(floating.Children) > 0 && floating.Children[len(floating.Children)-1].Kind == TextKind {
			floating.Children[len(floating.Children)-1].Data += node.Data
			continue
		}

		floating.Children = append(floating.Children, node)
	}
}

func (p *Parser) parse(t any) (*Node, bool, error) {
	switch token := t.(type) {
	case Text:
		return &Node{Kind: TextKind, Data: string(token)}, true, nil
	case Symbol:
		return &Node{Kind: TextKind, Data: symbol(string(token))}, true, nil
	case Command:
		return p.command(token)
	case Verbatim:
		return p.verbatim(token)
	case OptionalStart:
		return &Node{Kind: TextKind, Data: "["}, true, nil
	case OptionalEnd:
		return &Node{Kind: TextKind, Data: "]"}, true, nil
	case EnvironmentStart:
		return p.environment(token)
	case ParameterStart:
		// a bit of guessing here, this is hanging group it may enclose block or inline elements
		// we parse it as vertical layout and then try to figure it out
		children, _, err := p.vertical(func(a any, err error) bool {
			_, ok := a.(ParameterEnd)
			return err == nil && ok
		})

		if err != nil {
			return nil, false, err
		}

		// empty group
		if len(children) == 0 {
			return &Node{Kind: TextKind}, true, nil
		}

		if len(children) == 1 {
			node := children[0]

			// single paragraph means all items were text spans, return node as inline
			if node.Kind == ElementKind && node.Data == "\\par" {
				// check if it's group with a command, like this: {\cmd ...} and use \cmd to wrap group, so it looks like \cmd{...}
				if len(node.Children) != 0 {
					fc := node.Children[0]
					if fc.Kind == ElementKind && identifier.MatchString(fc.Data) && len(fc.Children) == 0 {
						return &Node{Kind: ElementKind, Data: fc.Data, Children: node.Children[1:]}, true, nil
					}
				}

				return &Node{Kind: ElementKind, Data: "{}", Children: node.Children}, true, nil
			}

			return children[0], false, nil
		}

		return &Node{Kind: ElementKind, Data: "{}", Children: children}, false, nil
	default:
		return nil, false, fmt.Errorf("unexpected token %T", t)
	}
}

func (p *Parser) command(c Command) (*Node, bool, error) {
	switch c {
	case "\\symbol":
		return p.symbol(c)
	case "\\par", "\\\\", "\\\\*", "\\newline", "\\dots", "\\ldots", "\\cdots", "\\vdots", "\\ddots", "\\InputFile", "\\InputData", "\\OutputFile", "\\Note", "\\Scoring", "\\Interaction", "\\Example", "\\Examples", "\\hskip", "\\vskip":
		return &Node{Kind: ElementKind, Data: string(c)}, true, nil
	case "\\underline", "\\emph", "\\sout", "\\textmd", "\\textbf", "\\textup", "\\textit", "\\textsl", "\\textsc", "\\textsf", "\\textrm", "\\bf", "\\it", "\\t", "\\tt", "\\texttt", "\\tiny", "\\scriptsize", "\\small", "\\normalsize", "\\large", "\\Large", "\\LARGE", "\\huge", "\\Huge", "\\section", "\\subsection", "\\bfseries", "\\itshape":
		return p.format(c)
	case "\\includegraphics":
		return p.graphics(c)
	case "\\url":
		return p.url(c)
	case "\\href":
		return p.href(c)
	case "\\def":
		return p.def(c)
	case "\\epigraph":
		return p.epigraph(c)
	case "\\vspace":
		return p.vspace(c)
	case "\\hspace":
		return p.hspace(c)
	case "\\exmp":
		return p.exmp(c)
	case "\\exmpfile":
		return p.exmpfile(c)
	default:
		if v, ok := p.defs[string(c)]; ok {
			return &Node{Kind: TextKind, Data: v}, true, nil
		}

		if v, ok := replacements[string(c)]; ok {
			return &Node{Kind: TextKind, Data: v}, true, nil
		}

		return nil, false, fmt.Errorf("unknown command %v", c)
	}
}

func (p *Parser) verbatim(v Verbatim) (*Node, bool, error) {
	switch v.Kind {
	case "$":
		return &Node{Kind: ElementKind, Data: "$", Children: []*Node{{Kind: TextKind, Data: v.Data}}}, true, nil
	case "$$":
		return &Node{Kind: ElementKind, Data: "$$", Children: []*Node{{Kind: TextKind, Data: v.Data}}}, false, nil
	case "%", "comment":
		return nil, false, nil
	case "\\verb", "\\verb*":
		return &Node{Kind: ElementKind, Data: v.Kind, Children: []*Node{{Kind: TextKind, Data: v.Data}}}, true, nil
	case "verbatim", "lstlisting":
		return &Node{Kind: ElementKind, Data: v.Kind, Children: []*Node{{Kind: TextKind, Data: v.Data}}}, false, nil
	default:
		return nil, false, fmt.Errorf("unknown verbatim \"%v\"", v.Kind)
	}
}

func (p *Parser) environment(e EnvironmentStart) (*Node, bool, error) {
	switch e.Name {
	case "center", "example":
		return p.division(e)
	case "itemize", "enumerate":
		return p.list(e)
	case "tabular":
		return p.tabular(e)
	case "problem":
		return p.problem(e)
	case "wrapfigure":
		return p.wrapfigure(e)
	default:
		return nil, true, fmt.Errorf("unknown environment %v", e.Name)
	}
}

// symbol is a \\symbol command
func (p *Parser) symbol(c Command) (*Node, bool, error) {
	val, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, err
	}

	code, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return nil, false, fmt.Errorf("symbol command must take an integer as parameter: %w", err)
	}

	return &Node{Kind: TextKind, Data: string([]rune{int32(code)})}, true, nil
}

// format is a command without parameters
func (p *Parser) format(c Command) (*Node, bool, error) {
	children, _, err := p.parameter()
	if err != nil {
		return nil, false, err
	}

	return &Node{Kind: ElementKind, Data: string(c), Children: children}, true, nil
}

// graphics reads \\includegraphics command
func (p *Parser) graphics(c Command) (*Node, bool, error) {
	params := map[string]string{}

	options, ok, err := p.optionVerbatim()
	if err != nil {
		return nil, false, err
	}

	if ok {
		params["options"] = options
	}

	src, ok, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, err
	}

	if ok {
		params["src"] = src
	}

	return &Node{Kind: ElementKind, Data: string(c), Parameters: params}, false, nil
}

// url reads \\url command
func (p *Parser) url(c Command) (*Node, bool, error) {
	href, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, err
	}

	return &Node{Kind: ElementKind, Data: string(c), Parameters: map[string]string{"href": href}}, true, nil
}

// href reads \\href command
func (p *Parser) href(c Command) (*Node, bool, error) {
	href, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, err
	}

	children, _, err := p.parameter()
	if err != nil {
		return nil, false, err
	}

	return &Node{Kind: ElementKind, Data: string(c), Parameters: map[string]string{"href": href}, Children: children}, true, nil
}

// def reads \\def command
func (p *Parser) def(c Command) (*Node, bool, error) {
	// def is followed by identifier (ie. command)
	token, err := p.tokens.Token()
	if err != nil {
		return nil, false, fmt.Errorf("unable to read def identifier: %w", err)
	}

	key, ok := token.(Command)
	if !ok || !identifier.MatchString(string(key)) {
		return nil, false, errors.New("def must be followed by identifier, for example: \\xyz, got ")
	}

	val, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid value in def: %w", err)
	}

	p.Define(string(key), val)

	return nil, false, nil
}

// epigraph reads \\epigraph command
func (p *Parser) epigraph(c Command) (*Node, bool, error) {
	text, _, err := p.parameter()
	if err != nil {
		return nil, false, fmt.Errorf("invalid epigraph text parameter: %w", err)
	}

	source, _, err := p.parameter()
	if err != nil {
		return nil, false, fmt.Errorf("invalid epigraph source parameter: %w", err)
	}

	node := &Node{Kind: ElementKind, Data: string(c), Children: []*Node{
		{Kind: ElementKind, Data: "\\epigraph:text", Children: text},
		{Kind: ElementKind, Data: "\\epigraph:source", Children: source},
	}}

	return node, false, nil
}

// vspace reads \\vspace command
func (p *Parser) vspace(c Command) (*Node, bool, error) {
	height, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid vspace parameter: %w", err)
	}

	return &Node{Kind: ElementKind, Data: string(c), Parameters: map[string]string{"height": height}}, false, nil
}

// hspace reads \\hspace command
func (p *Parser) hspace(c Command) (*Node, bool, error) {
	width, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid hspace parameter: %w", err)
	}

	return &Node{Kind: ElementKind, Data: string(c), Parameters: map[string]string{"width": width}}, false, nil
}

// exmp reads \\exmp command
func (p *Parser) exmp(c Command) (*Node, bool, error) {
	input, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid exmp input parameter: %w", err)
	}

	output, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid exmp output parameter: %w", err)
	}

	node := &Node{Kind: ElementKind, Data: string(c), Parameters: map[string]string{"input": input, "output": output}}
	return node, false, nil
}

// exmpfile reads \\exmpfile command
func (p *Parser) exmpfile(c Command) (*Node, bool, error) {
	input, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid exmpfile input parameter: %w", err)
	}

	output, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid exmpfile output parameter: %w", err)
	}

	name, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid exmpfile name parameter: %w", err)
	}

	node := &Node{Kind: ElementKind, Data: string(c), Parameters: map[string]string{"input": input, "output": output, "name": name}}
	return node, false, nil
}

// division reads an environment without any parameter or special content requirements
func (p *Parser) division(e EnvironmentStart) (*Node, bool, error) {
	children, _, err := p.vertical(func(a any, err error) bool {
		n, ok := a.(EnvironmentEnd)
		return err == nil && ok && n.Name == e.Name
	})

	if err != nil {
		return nil, false, err
	}

	return &Node{Kind: ElementKind, Data: e.Name, Children: children}, false, nil
}

// list reads an environment with multiple items defined by \\item command
func (p *Parser) list(e EnvironmentStart) (*Node, bool, error) {
	var items []*Node
	itimized := false

	for {
		children, last, err := p.vertical(func(a any, err error) bool {
			if err != nil {
				return false
			}

			if n, ok := a.(EnvironmentEnd); ok {
				return n.Name == e.Name
			}

			if c, ok := a.(Command); ok {
				return string(c) == "\\item"
			}

			return false
		})

		if err != nil {
			return nil, false, err
		}

		if itimized {
			items = append(items, &Node{Kind: ElementKind, Data: "\\item", Children: children})
		}

		// this skip content until we found first \\item
		itimized = true

		if _, ok := last.(EnvironmentEnd); ok {
			break
		}
	}

	return &Node{Kind: ElementKind, Data: e.Name, Children: items}, false, nil
}

// tabular reads tabular environment, where cells are separated by "&" and rows are separated by \\
func (p *Parser) tabular(e EnvironmentStart) (*Node, bool, error) {
	pos, _, err := p.optionString()
	if err != nil {
		return nil, false, fmt.Errorf("unable to read tabular environment [pos] parameter: %w", err)
	}

	colspec, _, err := p.parameterString()
	if err != nil {
		return nil, false, fmt.Errorf("unable to read tabular environment {colspec} parameter: %w", err)
	}

	var rows []*Node
	hanging := &Node{Kind: ElementKind, Data: "\\row"}

	addCell := func(nodes []*Node) {
		if len(nodes) > 0 {
			hanging.Children = append(hanging.Children, &Node{Kind: ElementKind, Data: "\\cell", Children: nodes})
		}
	}

	addHanging := func() {
		if len(hanging.Children) > 0 {
			rows = append(rows, hanging)
			hanging = &Node{Kind: ElementKind, Data: "\\row"}
		}
	}

	for {
		children, last, err := p.vertical(func(a any, err error) bool {
			if err != nil {
				return false
			}

			if n, ok := a.(EnvironmentEnd); ok {
				return n.Name == e.Name
			}

			if n, ok := a.(Symbol); ok {
				return n == "&"
			}

			if c, ok := a.(Command); ok {
				return isNewline(string(c)) || string(c) == "\\hline"
			}

			return false
		})

		if err != nil {
			return nil, false, err
		}

		// depending on how we stopped reading,
		if n, ok := last.(Symbol); ok && n == "&" {
			// stopped by "&", add new cell
			addCell(children)
			continue
		}

		if c, ok := last.(Command); ok {
			// stopped by newline, add new row
			if isNewline(string(c)) {
				addCell(children)
				addHanging()
				continue
			}

			// stopped by hline, override current row with hline and start a new row
			if string(c) == "\\hline" {
				addHanging()
				rows = append(rows, &Node{Kind: ElementKind, Data: "\\hline"})
				continue
			}
		}

		// stopped by environment end, exit
		if _, ok := last.(EnvironmentEnd); ok {
			addCell(children)
			addHanging()
			break
		}
	}

	params := map[string]string{"colspec": colspec}
	if pos != "" {
		params["pos"] = pos
	}

	return &Node{Kind: ElementKind, Parameters: params, Data: e.Name, Children: rows}, false, nil
}

// problem reads problem environment, a special environment used for formatting problems in computer science competitions
func (p *Parser) problem(e EnvironmentStart) (*Node, bool, error) {
	params := map[string]string{}

	keys := []string{"title", "input", "output", "time_limit", "memory_limit"}
	for index, key := range keys {
		val, ok, err := p.parameterVerbatim()
		if err != nil {
			return nil, false, fmt.Errorf("unable to read parameter #%d (%s) in problem environment: %w", index, key, err)
		}

		if !ok {
			break
		}

		params[key] = val
	}

	children, _, err := p.vertical(func(a any, err error) bool {
		n, ok := a.(EnvironmentEnd)
		return err == nil && ok && n.Name == e.Name
	})

	if err != nil {
		return nil, false, err
	}

	return &Node{Kind: ElementKind, Data: e.Name, Parameters: params, Children: children}, false, nil
}

func (p *Parser) wrapfigure(e EnvironmentStart) (*Node, bool, error) {
	lineheight, _, err := p.optionVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid wrapfigure lineheight parameter: %w", err)
	}

	position, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid wrapfigure position parameter: %w", err)
	}

	width, _, err := p.parameterVerbatim()
	if err != nil {
		return nil, false, fmt.Errorf("invalid wrapfigure width parameter: %w", err)
	}

	params := map[string]string{
		"position": position,
		"width":    width,
	}

	if lineheight != "" {
		params["lineheight"] = lineheight
	}

	children, _, err := p.vertical(func(a any, err error) bool {
		n, ok := a.(EnvironmentEnd)
		return err == nil && ok && n.Name == e.Name
	})

	if err != nil {
		return nil, false, err
	}

	return &Node{Kind: ElementKind, Data: e.Name, Parameters: params, Children: children}, false, nil
}

// option reads optional parameter (wrapped in []) if token "t" is optional parameter start.
// It returns t if there is no optional parameter, or next token after optional parameter
func (p *Parser) option() ([]*Node, bool, error) {
	char, err := p.tokens.Peek()
	if err == io.EOF {
		return nil, false, nil
	}

	if err != nil || char != '[' {
		return nil, false, err
	}

	open, err := p.tokens.Token()
	if err != nil {
		return nil, false, err
	}

	if _, ok := open.(OptionalStart); !ok {
		return nil, false, fmt.Errorf("expected optional group beginning, but got %T instead", open)
	}

	val, err := p.horizontal(func(a any, err error) bool {
		_, ok := a.(OptionalEnd)
		return err == nil && ok
	})

	return val, true, err
}

// optionVerbatim reads optional parameter in verbatim mode
func (p *Parser) optionVerbatim() (string, bool, error) {
	char, err := p.tokens.Peek()
	if err == io.EOF {
		return "", false, nil
	}

	if err != nil || char != '[' {
		return "", false, err
	}

	open, err := p.tokens.Token()
	if err != nil {
		return "", false, err
	}

	if _, ok := open.(OptionalStart); !ok {
		return "", false, fmt.Errorf("expected optional group beginning, but got %T instead", open)
	}

	val, err := p.tokens.Verbatim(func(r rune, err error) bool {
		return err == io.EOF || (err == nil && r == ']')
	})

	return val, true, err
}

// optionString reads optional parameter and transforms it to string
func (p *Parser) optionString() (str string, ok bool, err error) {
	val, ok, err := p.option()
	if !ok || err != nil {
		return "", ok, err
	}

	str, err = stringify(val)
	return
}

// parameter reads obligatory (wrapped in {}) parameter
func (p *Parser) parameter() (children []*Node, ok bool, err error) {
	if err := p.tokens.Skip(); err != nil {
		return nil, false, err
	}

	char, err := p.tokens.Peek()
	if err == io.EOF {
		return nil, false, nil
	}

	if err != nil || char != '{' {
		return nil, false, err
	}

	open, err := p.tokens.Token()
	if err != nil {
		return nil, false, err
	}

	if _, ok := open.(ParameterStart); !ok {
		return nil, false, fmt.Errorf("expected parameter group beginning, but got %T instead", open)
	}

	val, err := p.horizontal(func(a any, err error) bool {
		_, ok := a.(ParameterEnd)
		return err == nil && ok
	})

	return val, true, err
}

// parameterVerbatim reads obligatory parameter in verbatim mode
func (p *Parser) parameterVerbatim() (str string, ok bool, err error) {
	if err := p.tokens.Skip(); err != nil {
		return "", false, err
	}

	char, err := p.tokens.Peek()
	if err == io.EOF {
		return "", false, nil
	}

	if err != nil || char != '{' {
		return "", false, err
	}

	open, err := p.tokens.Token()
	if err != nil {
		return "", false, err
	}

	if _, ok := open.(ParameterStart); !ok {
		return "", false, fmt.Errorf("expected parameter group beginning, but got %T instead", open)
	}

	val, err := p.tokens.Verbatim(func(r rune, err error) bool {
		return err == io.EOF || (err == nil && r == '}')
	})

	return val, true, err
}

// parameterString reads obligatory parameter and transforms it to string
func (p *Parser) parameterString() (str string, ok bool, err error) {
	val, ok, err := p.parameter()
	if !ok || err != nil {
		return "", ok, err
	}

	str, err = stringify(val)
	return
}
