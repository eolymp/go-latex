package latex

import (
	"reflect"
	"testing"
)

func TestMeasure(t *testing.T) {
	tt := []struct {
		name  string
		input string
		value float32
		unit  string
	}{
		{name: "px", input: "131.02px", value: 131.02, unit: "px"},
		{name: "em", input: ".025em", value: .025, unit: "em"},
		{name: "negative float", input: "-.025em", value: -.025, unit: "em"},
		{name: "negative int", input: "-25em", value: -25, unit: "em"},
		{name: "%", input: "25%", value: 25, unit: "%"},
		{name: "\\textwidth", input: "0.25\\textwidth", value: 0.25, unit: "\\textwidth"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v, u, err := Measure(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			if v != tc.value {
				t.Errorf("Value does not match: want %v, got %v", tc.value, v)
			}

			if u != tc.unit {
				t.Errorf("Unit does not match: want %v, got %v", tc.unit, u)
			}
		})
	}
}

func TestKeyValue(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		output map[string]string
	}{
		{name: "one arg", input: "key=value", output: map[string]string{"key": "value"}},
		{name: "few arg", input: "scale=1.2, angle=45", output: map[string]string{"scale": "1.2", "angle": "45"}},
		{name: "lower case", input: "SCALE=1.2, angle=45", output: map[string]string{"scale": "1.2", "angle": "45"}},
		{name: "with spaces", input: "scale=1.2, angle=    45", output: map[string]string{"scale": "1.2", "angle": "45"}},
		{name: "escaped values", input: "escaped=\"scale=1.2, \\\"angle\\\"=    45\", another=44", output: map[string]string{"escaped": "scale=1.2, \"angle\"=    45", "another": "44"}},
		{name: "single-quote escaped values", input: "escaped='scale=1.2, \\'angle\\'=    45', another=44", output: map[string]string{"escaped": "scale=1.2, 'angle'=    45", "another": "44"}},
		{name: "escaped value followed by spaces", input: "a=\"1\"    , b=30", output: map[string]string{"a": "1", "b": "30"}},
		{name: "values surrounded by spaces", input: "a = 1 , b = 3", output: map[string]string{"a": "1", "b": "3"}},
		{name: "escaped value followed by more stuff", input: "a=\"1\"    23, b=30", output: map[string]string{"a": "123", "b": "30"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v := KeyValue(tc.input)

			if !reflect.DeepEqual(v, tc.output) {
				t.Errorf("Value does not match: want %v, got %v", tc.output, v)
			}
		})
	}
}
