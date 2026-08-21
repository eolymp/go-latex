package latex

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestKeyValue(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		output map[string]string
	}{
		{
			name:   "one arg",
			input:  "key=value",
			output: map[string]string{"key": "value"},
		},
		{
			name:   "few arg",
			input:  "scale=1.2, angle=45",
			output: map[string]string{"scale": "1.2", "angle": "45"},
		},
		{
			name:   "lower case",
			input:  "SCALE=1.2, angle=45",
			output: map[string]string{"scale": "1.2", "angle": "45"},
		},
		{
			name:   "with spaces",
			input:  "scale=1.2, angle=    45",
			output: map[string]string{"scale": "1.2", "angle": "45"},
		},
		{
			name:   "escaped values",
			input:  "escaped=\"scale=1.2, \\\"angle\\\"=    45\", another=44",
			output: map[string]string{"escaped": "scale=1.2, \"angle\"=    45", "another": "44"},
		},
		{
			name:   "single-quote escaped values",
			input:  "escaped='scale=1.2, \\'angle\\'=    45', another=44",
			output: map[string]string{"escaped": "scale=1.2, 'angle'=    45", "another": "44"},
		},
		{
			name:   "escaped value followed by spaces",
			input:  "a=\"1\"    , b=30",
			output: map[string]string{"a": "1", "b": "30"},
		},
		{
			name:   "values surrounded by spaces",
			input:  "a = 1 , b = 3",
			output: map[string]string{"a": "1", "b": "3"},
		},
		{
			name:   "cyrillic values",
			input:  "type=note, title=\"Привіт 👋\"",
			output: map[string]string{"type": "note", "title": "Привіт 👋"},
		},
		{
			name:   "ignore invalid parts",
			input:  "type=note @ 2, fo, from=hello@eolymp.com",
			output: map[string]string{"type": "note", "from": "hello@eolymp.com"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			v, err := KeyValue(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(v, tc.output) {
				t.Errorf("Value does not match:\n%s\n", cmp.Diff(tc.output, v))
			}
		})
	}
}
