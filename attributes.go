package latex

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var measure = regexp.MustCompile("^(-?[0-9]*(?:\\.[0-9]+)?)(%|\\\\?[a-z ]*)$")
var whitespaces = regexp.MustCompile("[ \n\t\r]+")

type keyValueParserState int

const (
	parsingKey keyValueParserState = iota
	parsingValue
)

// KeyValue parses key-value parameters in this format: key=value, key=value, for example as used in \\includegraphics option parameter.
func KeyValue(raw string) map[string]string {
	kv := map[string]string{}

	key := ""
	value := ""
	escape := rune(0)
	state := parsingKey

	for idx := 0; idx < len(raw); idx++ {
		char := rune(raw[idx])
		next := rune(0)
		if idx+1 < len(raw) {
			next = rune(raw[idx+1])
		}

		switch state {
		case parsingKey:
			if (char == ' ' || char == '\t') && key == "" { // skip leading key whitespaces
				continue
			}

			if char == ',' { // key without value, skipping
				key = ""
				continue
			}

			if char == '=' {
				state = parsingValue
				continue
			}

			key += string(char)
		case parsingValue:
			// reached escape final char
			if escape != 0 && char == escape {
				for ; idx < len(raw)-1; idx++ { // skip all following whitespaces
					if raw[idx+1] != ' ' {
						break
					}
				}

				escape = 0
				continue
			}

			if escape == 0 && char == ',' {
				kv[strings.TrimSpace(strings.ToLower(key))] = strings.TrimSpace(value)
				state = parsingKey
				escape = 0
				key = ""
				value = ""
				continue
			}

			if value == "" && char == ' ' {
				continue
			}

			if value == "" && (char == '"' || char == '\'') {
				escape = char
				continue
			}

			if escape != 0 && char == '\\' && (next == '\\' || next == '"' || next == escape) {
				value += string(next)
				idx++
				continue
			}

			value += string(char)
		}
	}

	if state == parsingValue && escape == 0 {
		kv[strings.TrimSpace(strings.ToLower(key))] = strings.TrimSpace(value)
	}

	return kv
}

type ColumnSpec struct {
	BorderLeft  bool   // column should have left border
	BorderRight bool   // column should have right border
	Align       string // column alignment: c, l or r
}

// ColumnSpecs parses column spec in tabular environment
// todo: add support for repeated syntax *{x}{...}
// todo: if not support, at least correctly handle @{} and !{}
func ColumnSpecs(raw string) (spec []ColumnSpec) {
	raw = whitespaces.ReplaceAllString(raw, "") // remove all spaces since they don't have any meaning
	for pos, char := range raw {
		if char == '|' {
			continue
		}

		if char == 'c' || char == 'l' || char == 'r' {
			spec = append(spec, ColumnSpec{
				BorderLeft:  pos > 0 && raw[pos-1] == '|',
				BorderRight: pos < len(raw)-1 && raw[pos+1] == '|',
				Align:       string([]rune{char}),
			})
		}
	}

	return
}

// Measure parses measurement value, a number and units, for example: 5.1cm, 6em, 0.25\textwidth
func Measure(raw string) (float32, string, error) {
	match := measure.FindStringSubmatch(raw)
	if len(match) == 0 {
		return 0, "", errors.New("unable to parse measurement")
	}

	number, err := strconv.ParseFloat(match[1], 32)
	if err != nil {
		return 0, "", err
	}

	return float32(number), match[2], nil
}

func MeasurePixels(raw string) (float32, error) {
	n, u, err := Measure(raw)
	if err != nil {
		return 0, err
	}

	return ToPixels(n, u)
}

func ToPixels(value float32, unit string) (float32, error) {
	switch unit {
	case "pt":
		return float32(value) * cmInPixel / 28.4495, nil
	case "mm":
		return float32(value) * cmInPixel / 10, nil
	case "cm":
		return float32(value) * cmInPixel, nil
	case "in":
		return float32(value) * cmInPixel * 2.54, nil
	case "ex":
		return float32(value) * cmInPixel * 0.15132, nil
	case "em":
		return float32(value) * cmInPixel * 0.35146, nil
	case "px":
		return value, nil
	default:
		return 0, fmt.Errorf("measurement unit %#v is not supported", unit)
	}
}
