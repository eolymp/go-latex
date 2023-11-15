package latex_test

import (
	"github.com/eolymp/go-latex"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	tt := []struct {
		name   string
		input  string
		output []any
	}{
		{
			name:  "text",
			input: "one\ntwo\nthree",
			output: []any{
				latex.Text("one\n"),
				latex.Text("two\n"),
				latex.Text("three"),
			},
		},
		{
			name:  "command",
			input: "\\textbf{foo\\par bar}",
			output: []any{
				latex.Command("\\textbf"),
				latex.ParameterStart{},
				latex.Text("foo"),
				latex.Command("\\par"),
				latex.Text("bar"),
				latex.ParameterEnd{},
			},
		},
		{
			name:  "math",
			input: "foo $a_i^2 + b_i^2 \\le a_{i+1}^2$ bar",
			output: []any{
				latex.Text("foo "),
				latex.Verbatim{Kind: "$", Data: "a_i^2 + b_i^2 \\le a_{i+1}^2"},
				latex.Text(" bar"),
			},
		},
		{
			name:  "math with escaped $ symbol",
			input: "foo $a_i^2 + b_i^2 \\$ \\le a_{i+1}^2$ bar",
			output: []any{
				latex.Text("foo "),
				latex.Verbatim{Kind: "$", Data: "a_i^2 + b_i^2 \\$ \\le a_{i+1}^2"},
				latex.Text(" bar"),
			},
		},
		{
			name:  "math block",
			input: "foo $$a_i^2 + b_i^2 \\le a_{i+1}^2$$ bar",
			output: []any{
				latex.Text("foo "),
				latex.Verbatim{Kind: "$$", Data: "a_i^2 + b_i^2 \\le a_{i+1}^2"},
				latex.Text(" bar"),
			},
		},
		{
			name:  "math block with escaped $ symbols",
			input: "foo $$a_i^2 + b_i^2 \\$$ \\le a_{i+1}^2$$ bar",
			output: []any{
				latex.Text("foo "),
				latex.Verbatim{Kind: "$$", Data: "a_i^2 + b_i^2 \\$$ \\le a_{i+1}^2"},
				latex.Text(" bar"),
			},
		},
		{
			name:  "math block with $ inside (invalid formatting)",
			input: "foo $$a_i^2 + b_i^2 $ \\le a_{i+1}^2$$ bar",
			output: []any{
				latex.Text("foo "),
				latex.Verbatim{Kind: "$$", Data: "a_i^2 + b_i^2 $ \\le a_{i+1}^2"},
				latex.Text(" bar"),
			},
		},
		{
			name:  "optional group",
			input: "\\includegraphics[scale=1.5]{eolymp.png}",
			output: []any{
				latex.Command("\\includegraphics"),
				latex.OptionalStart{},
				latex.Text("scale=1.5"),
				latex.OptionalEnd{},
				latex.ParameterStart{},
				latex.Text("eolymp.png"),
				latex.ParameterEnd{},
			},
		},
		{
			name:  "oneline comment",
			input: "one\ntwo%comment\\foo\nthree",
			output: []any{
				latex.Text("one\n"),
				latex.Text("two"),
				latex.Verbatim{Kind: "%", Data: "comment\\foo"},
				latex.Text("three"),
			},
		},
		{
			name:  "block comment",
			input: "a\\begin{comment}This is\n multiline\ncomment\n\\end{comment}z",
			output: []any{
				latex.Text("a"),
				latex.EnvironmentStart{Name: "comment"},
				latex.Text("This is\n"),
				latex.Text(" multiline\n"),
				latex.Text("comment\n"),
				latex.EnvironmentEnd{Name: "comment"},
				latex.Text("z"),
			},
		},
		{
			name:  "verb command",
			input: "The \\verb|\\ldots| command \\ldots",
			output: []any{
				latex.Text("The "),
				latex.Verbatim{Kind: "\\verb", Data: "\\ldots", Attr: map[string]string{"delimiter": "|"}},
				latex.Text(" command "),
				latex.Command("\\ldots"),
			},
		},
		{
			name:  "verbatim environment",
			input: "\\begin{verbatim}\n10 PRINT \"HELLO WORLD \";\n20 GOTO 10\n\\end{verbatim}",
			output: []any{
				latex.EnvironmentStart{Name: "verbatim"},
				latex.Text("\n"),
				latex.Text("10 PRINT \"HELLO WORLD \";\n"),
				latex.Text("20 GOTO 10\n"),
				latex.EnvironmentEnd{Name: "verbatim"},
			},
		},
		{
			name:  "verb command with star",
			input: "\\verb*|like   this :-) |",
			output: []any{
				latex.Verbatim{Kind: "\\verb*", Data: "like   this :-) ", Attr: map[string]string{"delimiter": "|"}},
			},
		},
		{
			name:  "cf1",
			input: "These are inline formulas: $x$, $a_i^2 + b_i^2 \\le a_{i+1}^2$. Afterwards...",
			output: []any{
				latex.Text("These are inline formulas: "),
				latex.Verbatim{Kind: "$", Data: "x"},
				latex.Text(", "),
				latex.Verbatim{Kind: "$", Data: "a_i^2 + b_i^2 \\le a_{i+1}^2"},
				latex.Text(". Afterwards..."),
			},
		},
		{
			name:  "cf2",
			input: "These are centered formulas: $$x,$$ $$a_i^2 + b_i^2 \\le a_{i+1}^2.$$ Afterwards...",
			output: []any{
				latex.Text("These are centered formulas: "),
				latex.Verbatim{Kind: "$$", Data: "x,"},
				latex.Text(" "),
				latex.Verbatim{Kind: "$$", Data: "a_i^2 + b_i^2 \\le a_{i+1}^2."},
				latex.Text(" Afterwards..."),
			},
		},
		{
			name:  "cf3",
			input: "Some complex formula: $$P(|S - E[S]| \\ge t) \\le 2 \\exp \\left( -\\frac{2 t^2 n^2}{\\sum_{i = 1}^n (b_i - a_i)^2} \\right).$$",
			output: []any{
				latex.Text("Some complex formula: "),
				latex.Verbatim{Kind: "$$", Data: "P(|S - E[S]| \\ge t) \\le 2 \\exp \\left( -\\frac{2 t^2 n^2}{\\sum_{i = 1}^n (b_i - a_i)^2} \\right)."},
			},
		},
		{
			name:  "cf4",
			input: "First paragraph.\nStill first paragraph.",
			output: []any{
				latex.Text("First paragraph.\n"),
				latex.Text("Still first paragraph."),
			},
		},
		{
			name:  "cf5",
			input: "First paragraph.\n\nSecond paragraph.",
			output: []any{
				latex.Text("First paragraph.\n"),
				latex.Text("\n"),
				latex.Text("Second paragraph."),
			},
		},
		{
			name:  "cf6",
			input: "\\bf{This text is bold.}or\n\\textbf{This text is bold.}",
			output: []any{
				latex.Command("\\bf"),
				latex.ParameterStart{},
				latex.Text("This text is bold."),
				latex.ParameterEnd{},
				latex.Text("or\n"),
				latex.Command("\\textbf"),
				latex.ParameterStart{},
				latex.Text("This text is bold."),
				latex.ParameterEnd{},
			},
		},
		{
			name:  "cf7",
			input: "\\t{This text is monospaced.}",
			output: []any{
				latex.Command("\\t"),
				latex.ParameterStart{},
				latex.Text("This text is monospaced."),
				latex.ParameterEnd{},
			},
		},
		{
			name:  "cf8",
			input: "\\Huge{This text is 200\\% of normal size.}",
			output: []any{
				latex.Command("\\Huge"),
				latex.ParameterStart{},
				latex.Text("This text is 200"),
				latex.Text("%"),
				latex.Text(" of normal size."),
				latex.ParameterEnd{},
			},
		},
		{
			name:  "cf9",
			input: "\\begin{center}\n  \\def \\htmlPixelsInCm {45}  % pixels in 1 centimeter in HTML mode\n  \\includegraphics[width=4cm]{logo.png} \\\\\n  \\small{Centered image with width specified (180px).}\n\\end{center}",
			output: []any{
				latex.EnvironmentStart{Name: "center"},
				latex.Text("\n"),
				latex.Text("  "),
				latex.Command("\\def"),
				latex.Command("\\htmlPixelsInCm"),
				latex.ParameterStart{},
				latex.Text("45"),
				latex.ParameterEnd{},
				latex.Text("  "),
				latex.Verbatim{Kind: "%", Data: " pixels in 1 centimeter in HTML mode"},
				latex.Command("\\includegraphics"),
				latex.OptionalStart{},
				latex.Text("width=4cm"),
				latex.OptionalEnd{},
				latex.ParameterStart{},
				latex.Text("logo.png"),
				latex.ParameterEnd{},
				latex.Text(" "),
				latex.Command("\\\\"),
				latex.Command("\\small"),
				latex.ParameterStart{},
				latex.Text("Centered image with width specified (180px)."),
				latex.ParameterEnd{},
				latex.Text("\n"),
				latex.EnvironmentEnd{Name: "center"},
			},
		},
		{
			name:  "cf10",
			input: "This is the unordered list:\n\\begin{itemize}\n  \\item This is the first item;\n  \\item This is the second item.\n\\end{itemize}",
			output: []any{
				latex.Text("This is the unordered list:\n"),
				latex.EnvironmentStart{Name: "itemize"},
				latex.Text("\n"),
				latex.Text("  "),
				latex.Command("\\item"),
				latex.Text("This is the first item;\n"),
				latex.Text("  "),
				latex.Command("\\item"),
				latex.Text("This is the second item.\n"),
				latex.EnvironmentEnd{Name: "itemize"},
			},
		},
		{
			name:  "cf11",
			input: "If you want to quote single character, use single quotes: `a'.\n\nIn some statements use <<these double quotes>>. As for the long dashes~--- use these like that.\n\nIn English statements use ``these double quotes''. As for the long dashes \"--- use these like that.",
			output: []any{
				latex.Text("If you want to quote single character, use single quotes: "),
				latex.Symbol("`"),
				latex.Text("a"),
				latex.Symbol("'"),
				latex.Text(".\n"),
				latex.Text("\n"),
				latex.Text("In some statements use "),
				latex.Symbol("<<"),
				latex.Text("these double quotes"),
				latex.Symbol(">>"),
				latex.Text(". As for the long dashes"),
				latex.Symbol("~"),
				latex.Symbol("---"),
				latex.Text(" use these like that.\n"),
				latex.Text("\n"),
				latex.Text("In English statements use "),
				latex.Symbol("``"),
				latex.Text("these double quotes"),
				latex.Symbol("''"),
				latex.Text(". As for the long dashes \""),
				latex.Symbol("---"),
				latex.Text(" use these like that."),
			},
		},
		{
			name:  "char with one dec",
			input: "Bo\\char9",
			output: []any{
				latex.Text("Bo"),
				latex.Symbol("\t"),
			},
		},
		{
			name:  "char with two dec",
			input: "Bo\\char98",
			output: []any{
				latex.Text("Bo"),
				latex.Symbol("b"),
			},
		},
		{
			name:  "char with three dec",
			input: "Mo\\char101",
			output: []any{
				latex.Text("Mo"),
				latex.Symbol("e"),
			},
		},
		{
			name:  "char with oct",
			input: "What\\char'77",
			output: []any{
				latex.Text("What"),
				latex.Symbol("?"),
			},
		},
		{
			name:  "char with hex",
			input: "Mo\\char\"F0",
			output: []any{
				latex.Text("Mo"),
				latex.Symbol("ð"),
			},
		},
		{
			name:  "char with whitespaces",
			input: "\\texttt{\\char 94}",
			output: []any{
				latex.Command("\\texttt"),
				latex.ParameterStart{},
				latex.Symbol("^"),
				latex.ParameterEnd{},
			},
		},
		{
			name:  "single quotes are not replaced",
			input: "text < other text",
			output: []any{
				latex.Text("text "),
				latex.Text("<"),
				latex.Text(" other text"),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			lexer := latex.NewTokenizer(strings.NewReader(tc.input))

			var got []any

			for {
				token, err := lexer.Token()
				if err == io.EOF {
					break
				}

				if err != nil {
					t.Fatalf("Unable to read token: %v", err)
				}

				got = append(got, token)
			}

			want := tc.output

			if !reflect.DeepEqual(want, got) {
				t.Errorf("Tokens do not match:\n want %#v\n  got %#v\n", want, got)
			}
		})
	}
}
