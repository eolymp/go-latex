package latex

import (
	"errors"
	"fmt"
	"strconv"
)

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
