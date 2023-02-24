package latex

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// KeyValue parses key-value parameters in this format: key=value, key=value, for example as used in \\includegraphics option parameter.
func KeyValue(raw string) map[string]string {
	kv := map[string]string{}

	parts := strings.Split(raw, ",")
	for _, part := range parts {
		n := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(n) == 1 {
			kv[strings.ToLower(n[0])] = ""
			continue
		}

		if len(n) >= 2 {
			kv[strings.ToLower(n[0])] = n[1]
		}
	}

	return kv
}

var measure = regexp.MustCompile("^(-?[0-9]*(?:\\.[0-9]+)?)(%|\\\\?[a-z ]*)$")

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
