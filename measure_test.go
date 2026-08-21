package latex

import "testing"

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
