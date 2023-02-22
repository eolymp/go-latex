package latex

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Tokenizer struct {
	r io.RuneScanner
}

func NewTokenizer(r io.RuneScanner) *Tokenizer {
	return &Tokenizer{r: r}
}

func (l *Tokenizer) Token() (any, error) {
	char, _, err := l.r.ReadRune()
	if err != nil {
		return nil, err
	}

	switch char {
	case '{':
		return ParameterStart{}, nil
	case '}':
		return ParameterEnd{}, nil
	case '[':
		return OptionalStart{}, nil
	case ']':
		return OptionalEnd{}, nil
	case '&', '~', '#', '^', '_':
		return Symbol([]rune{char}), nil
	case '`', '\'', '-', '<', '>':
		return l.readLigature(char)
	case '%':
		return l.readLineComment()
	case '$':
		return l.readMath()
	case '\\':
		return l.readBackslash()
	default:
		if isSpecial(char) {
			return nil, fmt.Errorf("trying to read special char %c as text", char)
		}

		if err := l.r.UnreadRune(); err != nil {
			return nil, err
		}

		return l.readText()
	}
}

// Verbatim reads input rune by rune until stop returns true
func (l *Tokenizer) Verbatim(stop func(rune, error) bool) (string, error) {
	var runes []rune
	for {
		read, _, err := l.r.ReadRune()
		if stop(read, err) {
			return string(runes), nil
		}

		if err != nil {
			return "", err
		}

		runes = append(runes, read)
	}
}

func (l *Tokenizer) Peek() (rune, error) {
	read, _, err := l.r.ReadRune()
	if err != nil {
		return 0, err
	}

	return read, l.r.UnreadRune()
}

func (l *Tokenizer) readText() (any, error) {
	var runes []rune
	for {
		read, _, err := l.r.ReadRune()
		if err == io.EOF {
			return Text(runes), nil
		}

		if err != nil {
			return nil, err
		}

		if isSpecial(read) {
			return Text(runes), l.r.UnreadRune()
		}

		runes = append(runes, read)

		if read == '\n' {
			return Text(runes), nil
		}
	}
}

func (l *Tokenizer) readMath() (any, error) {
	// we already entered math with one $, check if next one is $ too (ie. math block)
	read, _, err := l.r.ReadRune()
	if err != nil {
		return "", err
	}

	isBlock := read == '$' // math is described in block (two $$ in the beginning and in the end)
	isClosing := false     // we found first closing $ for block and expecting one more

	var runes = []rune{'$', read}

	for {
		read, _, err := l.r.ReadRune()
		if err == io.EOF {
			return nil, errors.New("EOF: math block is not closed")
		}

		if err != nil {
			return nil, err
		}

		if read == '$' && runes[len(runes)-1] != '\\' {
			if !isBlock {
				return Verbatim{Kind: "$", Data: string(runes[1:])}, nil
			}

			if isClosing {
				return Verbatim{Kind: "$$", Data: string(runes[2:])}, nil
			}

			isClosing = true
			continue
		}

		// previous rune was $, but this one is not, so let's add $ because it's not part of the closing sequence
		if isClosing {
			runes = append(runes, '$')
		}

		isClosing = false
		runes = append(runes, read)
	}
}

func (l *Tokenizer) readBackslash() (any, error) {
	r, _, err := l.r.ReadRune()
	if err != nil {
		return nil, err
	}

	// one symbol command
	if isCommand(r) {
		star, err := l.star()
		if err != nil {
			return nil, err
		}

		if star {
			return Command([]rune{'\\', r, '*'}), l.Skip()
		}

		return Command([]rune{'\\', r}), l.Skip()
	}

	// a letter means it's a named command \xyz
	if isLetter(r) {
		if err := l.r.UnreadRune(); err != nil {
			return nil, err
		}

		return l.readCommand('\\')
	}

	// special character escaped by \\
	return Text(r), nil
}

func (l *Tokenizer) readCommand(start rune) (any, error) {
	runes := []rune{start}
	for {
		read, _, err := l.r.ReadRune()
		if err != io.EOF {
			if err != nil {
				return "", err
			}

			// letter: continue reading name
			if isLetter(read) {
				runes = append(runes, read)
				continue
			}

			// command names may include * in the end (except for begin and end)
			if read == '*' && string(runes) != "\\begin" && string(runes) != "\\end" {
				runes = append(runes, read)
			} else {
				if err := l.r.UnreadRune(); err != nil {
					return nil, err
				}
			}
		}

		command := string(runes)

		switch command {
		case "\\verb", "\\verb*":
			return l.readVerbatim(command)
		case "\\begin":
			return l.readBlockStart()
		case "\\end":
			return l.readBlockEnd()
		case "\\char":
			return l.readChar()
		default:
			if err := l.Skip(); err != nil {
				return nil, err
			}

			return Command(command), nil
		}
	}
}

func (l *Tokenizer) readBlockStart() (any, error) {
	if err := l.forwardTo('{'); err != nil {
		return nil, err
	}

	word, err := l.word()
	if err != nil {
		return nil, err
	}

	if word == "" {
		return nil, errors.New("environment name is expected")
	}

	if err := l.expect('}'); err != nil {
		return nil, err
	}

	if word == "comment" || word == "lstlisting" || word == "verbatim" {
		return l.readVerbatimBlock(word)
	}

	return EnvironmentStart{Name: word}, nil
}

func (l *Tokenizer) readBlockEnd() (any, error) {
	if err := l.forwardTo('{'); err != nil {
		return nil, err
	}

	word, err := l.word()
	if err != nil {
		return nil, err
	}

	if word == "" {
		return nil, errors.New("environment name is expected")
	}

	if err := l.expect('}'); err != nil {
		return nil, err
	}

	return EnvironmentEnd{Name: word}, nil
}

func (l *Tokenizer) readChar() (any, error) {
	if err := l.Skip(); err != nil {
		return nil, err
	}

	first, _, err := l.r.ReadRune()
	if err != nil {
		return nil, err
	}

	// char with dec code: \\char98
	if isDigit(first, 10) {
		if err := l.r.UnreadRune(); err != nil {
			return nil, err
		}

		number, err := l.readNumber(10)
		if err != nil {
			return nil, fmt.Errorf("\\char must be followed by exactly two digits: %w", err)
		}

		return Symbol([]rune{rune(number)}), nil
	}

	// char with oct code: \\char'77
	if first == '\'' {
		number, err := l.readNumber(8)
		if err != nil {
			return nil, fmt.Errorf("\\char must be followed by oct digits: %w", err)
		}

		return Symbol([]rune{rune(number)}), nil
	}

	// char with hex code: \\char"FF
	if first == '"' {
		number, err := l.readNumber(16)
		if err != nil {
			return nil, fmt.Errorf("\\char\" must be followed by hex digits: %w", err)
		}

		return Symbol([]rune{rune(number)}), nil
	}

	return nil, errors.New("\\char must be followed by a digit, ' or \"")
}

func (l *Tokenizer) readNumber(base int) (n int64, err error) {
	var buffer []rune
	for {
		read, _, err := l.r.ReadRune()
		if err == io.EOF {
			return strconv.ParseInt(string(buffer), base, 32)
		}
		if err != nil {
			return 0, err
		}

		if !isDigit(read, base) {
			if err := l.r.UnreadRune(); err != nil {
				return 0, err
			}

			return strconv.ParseInt(string(buffer), base, 32)
		}

		buffer = append(buffer, read)
	}
}

// readLineComment reads one line comment after %
//
// When LATEX encounters a % character while processing an input file, it ignores the
// rest of the present line, the line break, and all whitespace at the
// beginning of the next line.
func (l *Tokenizer) readLineComment() (any, error) {
	var runes []rune
	for {
		read, _, err := l.r.ReadRune()
		if err == io.EOF || read == '\n' {
			if err := l.Skip(); err != nil {
				return nil, err
			}

			return Verbatim{Kind: "%", Data: string(runes)}, nil
		}

		if err != nil {
			return nil, err
		}

		runes = append(runes, read)
	}
}

func (l *Tokenizer) readLigature(first rune) (any, error) {
	line := []rune{first}
	for {
		read, _, err := l.r.ReadRune()
		if err == io.EOF {
			return Symbol(line), nil
		}

		if err != nil {
			return nil, err
		}

		switch string(append(line, read)) {
		case "<<", ">>", "``", "--", "---", "''":
			line = append(line, read)
		default:
			return Symbol(line), l.r.UnreadRune()
		}
	}
}

// readVerbatimBlock reads verbatim block (ie. block where all markup is ignored) of a given type (eg. comment, verbatim etc)
// until it finds closing \\end command.
func (l *Tokenizer) readVerbatimBlock(kind string) (any, error) {
	if err := l.Skip(); err != nil {
		return nil, err
	}

	var runes []rune
	for {
		read, _, err := l.r.ReadRune()
		if err == io.EOF {
			return Verbatim{Data: string(runes)}, nil
		}

		if err != nil {
			return nil, err
		}

		runes = append(runes, read)

		// todo: what if end is escaped: \\\\end{kind}?
		if strings.HasSuffix(string(runes), "\\end{"+kind+"}") {
			return Verbatim{Kind: kind, Data: strings.TrimSuffix(string(runes), "\\end{"+kind+"}")}, nil
		}
	}
}

func (l *Tokenizer) readVerbatim(command string) (any, error) {
	delimiter, _, err := l.r.ReadRune()
	if err != nil {
		return nil, err
	}

	if isWhitespace(delimiter) || isLetter(delimiter) || delimiter == '*' {
		return nil, fmt.Errorf("delimiter character \"%c\" is not allowed", delimiter)
	}

	var runes []rune
	for {
		read, _, err := l.r.ReadRune()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if read == delimiter || err == io.EOF {
			return Verbatim{Kind: command, Data: string(runes)}, nil
		}

		runes = append(runes, read)
	}

}

// Skip until next non-whitespace symbol
func (l *Tokenizer) Skip() error {
	for {
		r, _, err := l.r.ReadRune()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if !isWhitespace(r) {
			return l.r.UnreadRune()
		}
	}
}

// forwardTo skips whitespaces and makes sure next symbol is "e"
func (l *Tokenizer) forwardTo(e rune) error {
	if err := l.Skip(); err != nil {
		return err
	}

	return l.expect(e)
}

// expect verifies than following symbol is "e"
func (l *Tokenizer) expect(e rune) error {
	r, _, err := l.r.ReadRune()
	if err == io.EOF {
		return nil
	}

	if r != e {
		return fmt.Errorf("expected symbol %c, got %c instead", e, r)
	}

	return nil
}

// star reads following star symbol, if present
func (l *Tokenizer) star() (bool, error) {
	r, _, err := l.r.ReadRune()
	if err == io.EOF {
		return false, nil
	}

	if r == '*' {
		return true, nil
	}

	return false, l.r.UnreadRune()
}

// word reads sequence of letters
func (l *Tokenizer) word() (string, error) {
	var runes []rune
	for {
		read, _, err := l.r.ReadRune()
		if err == io.EOF {
			return string(runes), nil
		}

		if err != nil {
			return "", err
		}

		if !isLetter(read) {
			return string(runes), l.r.UnreadRune()
		}

		runes = append(runes, read)
	}
}

// isLetter returns true for a letter
func isLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

// isDigit returns true for a dec digit
func isDigit(r rune, base int) bool {
	if base <= 1 {
		return r == '0'
	}

	if base <= 10 {
		return '0' <= r && r < rune('0'+base)
	}

	if '0' <= r && r < '9' {
		return true
	}

	return 'A' <= r && r <= rune('A'+base-10)
}

// isSpacial returns true if a symbol has a special meaning and should interrupt text reading
func isSpecial(r rune) bool {
	switch r {
	case '#', '$', '%', '^', '&', '_', '{', '}', '~', '\\', '[', ']', '`', '\'', '-', '<', '>':
		return true
	default:
		return false
	}
}

func isWhitespace(r rune) bool {
	switch r {
	case ' ', '\n', '\t', '\r':
		return true
	default:
		return false
	}
}

// isCommand checks if symbol represents "one-symbol" command
func isCommand(r rune) bool {
	switch r {
	case '\\', '-':
		return true
	default:
		return false
	}
}
