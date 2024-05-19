package latex

import (
	"io"
	"regexp"
	"strings"
)

var measure = regexp.MustCompile("^(-?[0-9]*(?:\\.[0-9]+)?)(%|\\\\?[a-z ]*)$")
var whitespaces = regexp.MustCompile("[ \n\t\r]+")

type keyValueParserState int

const (
	stateLookingForKey keyValueParserState = iota
	stateReadingKey
	stateKeyRead
	stateLookingForValue
	stateReadingValue
	stateLookingForDelimiter
)

func KeyValue(raw string) (map[string]string, error) {
	read := strings.NewReader(raw)
	attr := map[string]string{}

	key := ""
	value := ""
	state := stateLookingForKey
	escape := rune(0)

	for {
		char, _, err := read.ReadRune()
		if err == io.EOF {
			if state == stateReadingValue {
				attr[key] = value
			}

			return attr, nil
		}

		if err != nil {
			return nil, err
		}

		switch state {
		case stateLookingForKey:
			if isValidAttributeNameChar(char) {
				state = stateReadingKey
				key = string(char)
			}

			continue
		case stateReadingKey:
			if isValidAttributeNameChar(char) {
				key += string(char)
				continue
			}

			switch {
			case char == ' ':
				state = stateKeyRead
			case char == '=':
				state = stateLookingForValue
			default:
				state = stateLookingForKey
				key = ""
			}

			continue
		case stateKeyRead:
			switch {
			case char == ' ':
			case char == '=':
				state = stateLookingForValue
			default:
				state = stateLookingForKey
				key = ""
			}
		case stateLookingForValue:
			if char == '"' || char == '\'' {
				value = ""
				state = stateReadingValue
				escape = char
			} else if char != ' ' {
				value = string(char)
				state = stateReadingValue
				escape = 0
			}

			continue
		case stateReadingValue:
			if escape == 0 && char == ' ' {
				attr[key] = value
				state = stateLookingForDelimiter
				continue
			}

			if escape == 0 && char == ',' {
				attr[key] = value
				state = stateLookingForKey
				continue
			}

			if escape != 0 && char == '\\' {
				next, _, err := read.ReadRune()
				if err == io.EOF {
					value += string(char)
					attr[key] = value
					return attr, nil
				}

				if err != nil {
					return nil, err
				}

				if next == '\\' || next == '"' || next == escape {
					value += string(next)
					continue
				}

				if err := read.UnreadRune(); err != nil {
					return nil, err
				}
			}

			if escape != 0 && char == escape {
				attr[key] = value
				state = stateLookingForDelimiter
				continue
			}

			value += string(char)

		case stateLookingForDelimiter:
			if char == ',' {
				state = stateLookingForKey
			}
		}
	}
}

func isValidAttributeNameChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == '-' || char == '_')
}
