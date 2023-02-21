package latex_test

import (
	"encoding/json"
	"github.com/eolymp/go-latex"
	"reflect"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	doc := func(children ...*latex.Node) *latex.Node {
		return &latex.Node{Kind: latex.DocumentKind, Children: children}
	}

	text := func(t string) *latex.Node {
		return &latex.Node{Kind: latex.TextKind, Data: t}
	}

	par := func(children ...*latex.Node) *latex.Node {
		return &latex.Node{Kind: latex.ElementKind, Data: "\\par", Children: children}
	}

	element := func(command string, children ...*latex.Node) *latex.Node {
		return &latex.Node{Kind: latex.ElementKind, Data: command, Children: children}
	}

	elementp := func(command string, params map[string]string, children ...*latex.Node) *latex.Node {
		return &latex.Node{Kind: latex.ElementKind, Data: command, Parameters: params, Children: children}
	}

	tt := []struct {
		name   string
		input  string
		output *latex.Node
	}{
		{
			name:   "simple paragraph",
			input:  "one two\nthree",
			output: doc(par(text("one two\nthree"))),
		},
		{
			name:  "two paragraphs separated by new lines",
			input: "one\ntwo\n\nthree\n\n\n\n\n\nfour",
			output: doc(
				par(text("one\ntwo\n")),
				par(text("three\n")),
				par(text("four")),
			),
		},
		{
			name:  "simple formatting",
			input: "odd \\textbf{foo bar} baz",
			output: doc(par(
				text("odd "),
				element("\\textbf", text("foo bar")),
				text(" baz"),
			)),
		},
		{
			name:  "nested formatting",
			input: "odd \\textbf{foo \\textit{bar}} baz",
			output: doc(par(
				text("odd "),
				element("\\textbf", text("foo "), element("\\textit", text("bar"))),
				text(" baz"),
			)),
		},
		{
			name:  "inline math",
			input: "$\\alpha + \\beta$",
			output: doc(par(
				element("$", text("\\alpha + \\beta")),
			)),
		},
		{
			name:   "block math",
			input:  "$$\\alpha + \\beta$$",
			output: doc(element("$$", text("\\alpha + \\beta"))),
		},
		{
			name:  "math",
			input: "foo $a_i^2 + b_i^2 \\le a_{i+1}^2$ bar",
			output: doc(par(
				text("foo "),
				element("$", text("a_i^2 + b_i^2 \\le a_{i+1}^2")),
				text(" bar"),
			)),
		},
		{
			name:  "math block",
			input: "foo $$a_i^2 + b_i^2 \\le a_{i+1}^2$$ bar",
			output: doc(
				par(text("foo ")),
				element("$$", text("a_i^2 + b_i^2 \\le a_{i+1}^2")),
				par(text(" bar")),
			),
		},
		{
			name:  "image",
			input: "This is an image \\includegraphics[scale=1.5]{eolymp.png} some text after...",
			output: doc(
				par(text("This is an image ")),
				elementp("\\includegraphics", map[string]string{
					"options": "scale=1.5",
					"src":     "eolymp.png",
				}),
				par(text(" some text after...")),
			),
		},
		{
			name:   "oneline comment",
			input:  "one\ntwo%comment\\foo\nthree",
			output: doc(par(text("one\ntwothree"))),
		},
		{
			name:   "block comment",
			input:  "bar\\begin{comment}comment\\foo\\end{comment}baz",
			output: doc(par(text("barbaz"))),
		},
		{
			name:  "verb command",
			input: "The \\verb|\\ldots| command \\ldots",
			output: doc(par(
				text("The "),
				element("\\verb", text("\\ldots")),
				text(" command "),
				element("\\ldots"),
			)),
		},
		{
			name:  "verbatim environment",
			input: "the \\begin{verbatim}\n10 PRINT \"HELLO WORLD \";\n20 GOTO 10\n\\end{verbatim} code",
			output: doc(
				par(text("the ")),
				element("verbatim", text("10 PRINT \"HELLO WORLD \";\n20 GOTO 10\n")),
				par(text(" code")),
			),
		},
		{
			name:   "verb command with star",
			input:  "\\verb*|like   this :-) |",
			output: doc(par(element("\\verb*", text("like   this :-) ")))),
		},
		{
			name:  "cf1",
			input: "These are inline formulas: $x$, $a_i^2 + b_i^2 \\le a_{i+1}^2$. Afterwards...",
			output: doc(par(
				text("These are inline formulas: "),
				element("$", text("x")),
				text(", "),
				element("$", text("a_i^2 + b_i^2 \\le a_{i+1}^2")),
				text(". Afterwards..."),
			)),
		},
		{
			name:  "cf2",
			input: "These are centered formulas: $$x,$$ $$a_i^2 + b_i^2 \\le a_{i+1}^2.$$ Afterwards...",
			output: doc(
				par(text("These are centered formulas: ")),
				element("$$", text("x,")),
				par(text(" ")),
				element("$$", text("a_i^2 + b_i^2 \\le a_{i+1}^2.")),
				par(text(" Afterwards...")),
			),
		},
		{
			name:  "cf3",
			input: "Some complex formula: $$P(|S - E[S]| \\ge t) \\le 2 \\exp \\left( -\\frac{2 t^2 n^2}{\\sum_{i = 1}^n (b_i - a_i)^2} \\right).$$",
			output: doc(
				par(text("Some complex formula: ")),
				element("$$", text("P(|S - E[S]| \\ge t) \\le 2 \\exp \\left( -\\frac{2 t^2 n^2}{\\sum_{i = 1}^n (b_i - a_i)^2} \\right).")),
			),
		},
		{
			name:   "cf4",
			input:  "First paragraph.\nStill first paragraph.",
			output: doc(par(text("First paragraph.\nStill first paragraph."))),
		},
		{
			name:  "cf5",
			input: "First paragraph.\n\nSecond paragraph.",
			output: doc(
				par(text("First paragraph.\n")),
				par(text("Second paragraph.")),
			),
		},
		{
			name:  "cf6",
			input: "\\bf{This text is bold.}or\n\\textbf{This text is bold.}",
			output: doc(par(
				element("\\bf", text("This text is bold.")),
				text("or\n"),
				element("\\textbf", text("This text is bold.")),
			)),
		},
		{
			name:  "cf7",
			input: "\\it{This text is italic.}or\n\\textit{This text is italic.}",
			output: doc(par(
				element("\\it", text("This text is italic.")),
				text("or\n"),
				element("\\textit", text("This text is italic.")),
			)),
		},
		{
			name:  "cf8",
			input: "\\t{This text is monospaced.}or\n\\tt{This text is monospaced.}or\n\\texttt{This text is monospaced.}",
			output: doc(par(
				element("\\t", text("This text is monospaced.")),
				text("or\n"),
				element("\\tt", text("This text is monospaced.")),
				text("or\n"),
				element("\\texttt", text("This text is monospaced.")),
			)),
		},
		{
			name:  "cf9",
			input: "\\emph{This text is underlined.}or\n\\underline{This text is underlined.}",
			output: doc(par(
				element("\\emph", text("This text is underlined.")),
				text("or\n"),
				element("\\underline", text("This text is underlined.")),
			)),
		},
		{
			name:   "cf10",
			input:  "\\sout{This text is struck out.}",
			output: doc(par(element("\\sout", text("This text is struck out.")))),
		},
		{
			name:   "cf11",
			input:  "\\textsc{This text is capitalized.}",
			output: doc(par(element("\\textsc", text("This text is capitalized.")))),
		},
		{
			name:   "cf12",
			input:  "\\tiny{This text is 70\\% of normal size.}",
			output: doc(par(element("\\tiny", text("This text is 70% of normal size.")))),
		},
		{
			name:   "cf13",
			input:  "\\scriptsize{This text is 75\\% of normal size.}",
			output: doc(par(element("\\scriptsize", text("This text is 75% of normal size.")))),
		},
		{
			name:   "cf14",
			input:  "\\small{This text is 85\\% of normal size.}",
			output: doc(par(element("\\small", text("This text is 85% of normal size.")))),
		},
		{
			name:  "cf15",
			input: "\\normalsize{This text is 100\\% of normal size.}\nor just\nThis text is 100\\% of normal size.",
			output: doc(par(
				element("\\normalsize", text("This text is 100% of normal size.")),
				text("\nor just\nThis text is 100% of normal size."),
			)),
		},
		{
			name:  "cf16..20",
			input: "\\large{This text is 115\\% of normal size.}\\Large{This text is 130\\% of normal size.}\\LARGE{This text is 145\\% of normal size.}\\huge{This text is 175\\% of normal size.}\\Huge{This text is 200\\% of normal size.}",
			output: doc(par(
				element("\\large", text("This text is 115% of normal size.")),
				element("\\Large", text("This text is 130% of normal size.")),
				element("\\LARGE", text("This text is 145% of normal size.")),
				element("\\huge", text("This text is 175% of normal size.")),
				element("\\Huge", text("This text is 200% of normal size.")),
			)),
		},
		{
			name:  "cf21",
			input: "This is the unordered list:\n\\begin{itemize}\n  \\item This is the first item;\n  \\item This is the second item.\n\\end{itemize}",
			output: doc(
				par(text("This is the unordered list:\n")),
				element("itemize",
					element("\\item", par(text("This is the first item;\n  "))),
					element("\\item", par(text("This is the second item.\n"))),
				),
			),
		},
		{
			name:  "cf22",
			input: "This is the ordered list:\n\\begin{enumerate}\n  \\item This is the first item;\n  \\item This is the second item.\n\\end{enumerate}",
			output: doc(
				par(text("This is the ordered list:\n")),
				element("enumerate",
					element("\\item", par(text("This is the first item;\n  "))),
					element("\\item", par(text("This is the second item.\n"))),
				),
			),
		},
		{
			name:  "cf23",
			input: "Some C++ source code (auto-detecting and highlighting):\n\\begin{lstlisting}\n#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n\\end{lstlisting}",
			output: doc(
				par(text("Some C++ source code (auto-detecting and highlighting):\n")),
				element("lstlisting", text("#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n")),
			),
		},
		{
			name:  "cf24",
			input: "Link to website:\n\\url{https://eolymp.com/}.",
			output: doc(par(
				text("Link to website:\n"),
				elementp("\\url", map[string]string{"href": "https://eolymp.com/"}),
				text("."),
			)),
		},
		{
			name:  "cf25",
			input: "Link to website with caption:\n\\href{https://eolymp.com/}{Eolymp}.",
			output: doc(par(
				text("Link to website with caption:\n"),
				elementp("\\href", map[string]string{"href": "https://eolymp.com/"}, text("Eolymp")),
				text("."),
			)),
		},
		{
			name:  "cf26",
			input: "\\begin{center}\n  This content is centered.\n\n  $abacaba$\n\\end{center}",
			output: doc(element("center",
				par(text("\n  This content is centered.\n")),
				par(text("  "), element("$", text("abacaba")), text("\n")),
			)),
		},
		{
			name:  "cf27",
			input: "Unscaled image:\n\\includegraphics{eolymp.png}",
			output: doc(
				par(text("Unscaled image:\n")),
				elementp("\\includegraphics", map[string]string{"src": "eolymp.png"}),
			),
		},
		{
			name:  "cf28",
			input: "\\begin{center}\n  \\includegraphics{eolymp.png} \\\\\n  \\small{Centered unscaled image.}\n\\end{center}",
			output: doc(element("center",
				par(text("\n  ")),
				elementp("\\includegraphics", map[string]string{"src": "eolymp.png"}),
				par(
					text(" "),
					element("\\\\"),
					element("\\small", text("Centered unscaled image.")),
					text("\n"),
				),
			)),
		},
		{
			name:  "cf29",
			input: "\\begin{center}\n  \\includegraphics[scale=1.5]{eolymp.png} \\\\\n  \\small{Centered scaled image.}\n\\end{center}",
			output: doc(element("center",
				par(text("\n  ")),
				elementp("\\includegraphics", map[string]string{"src": "eolymp.png", "options": "scale=1.5"}),
				par(
					text(" "),
					element("\\\\"),
					element("\\small", text("Centered scaled image.")),
					text("\n"),
				),
			)),
		},
		{
			name:  "cf30",
			input: "\\begin{center}\n  \\def \\htmlPixelsInCm {45}  % pixels in 1 centimeter in HTML mode\n  \\includegraphics[width=4cm]{eolymp.png} \\\\\n  \\small{Centered image with width specified (180px).}\n\\end{center}",
			output: doc(element("center",
				par(text("\n    ")),
				elementp("\\includegraphics", map[string]string{"src": "eolymp.png", "options": "width=4cm"}),
				par(
					text(" "),
					element("\\\\"),
					element("\\small", text("Centered image with width specified (180px).")),
					text("\n"),
				),
			)),
		},
		{
			name:  "cf31",
			input: "Simple table without borders:\n\\begin{tabular}{ll}\n  First & Second \\\\\n  Third & Fourth\n\\end{tabular}",
			output: doc(
				par(text("Simple table without borders:\n")),
				elementp("tabular", map[string]string{"colspec": "ll"},
					element("\\row",
						element("\\cell", par(text("\n  First "))),
						element("\\cell", par(text(" Second "))),
					),
					element("\\row",
						element("\\cell", par(text("Third "))),
						element("\\cell", par(text(" Fourth\n"))),
					),
				),
			),
		},
		{
			name:  "cf32",
			input: "More complex table with borders:\n\\begin{tabular}{|l|c|r|} \\hline\n  Left aligned column & Centered column & Right aligned column \\\\ \\hline\n  Text & Text & Text \\\\ \\hline\n\\end{tabular}",
			output: doc(
				par(text("More complex table with borders:\n")),
				elementp("tabular", map[string]string{"colspec": "|l|c|r|"},
					element("\\hline"),
					element("\\row",
						element("\\cell", par(text("Left aligned column "))),
						element("\\cell", par(text(" Centered column "))),
						element("\\cell", par(text(" Right aligned column "))),
					),
					element("\\hline"),
					element("\\row",
						element("\\cell", par(text("Text "))),
						element("\\cell", par(text(" Text "))),
						element("\\cell", par(text(" Text "))),
					),
					element("\\hline"),
				),
			),
		},
		{
			name:  "cf33",
			input: "Scoring table example:\n\\begin{center}\n  \\begin{tabular}{ | c | c | c | c | } \\hline\n    \\bf{Group} &\n    \\bf{Add. constraints} &\n    \\bf{Points} &\n    \\bf{Req. groups} \\\\ \\hline\n    $1$ & $b = a + 1$ & $30$ & --- \\\\ \\hline\n    $2$ & $n \\le 1\\,000$ & $10$ & examples \\\\ \\hline\n    $3$ & $n \\le 10^7$ & $20$ & $2$ \\\\ \\hline\n    $4$ & --- & $40$ & $1$, $3$ \\\\ \\hline\n  \\end{tabular}\n\\end{center}",
			output: doc(
				par(text("Scoring table example:\n")),
				element("center",
					par(text("\n  ")),
					elementp("tabular", map[string]string{"colspec": " | c | c | c | c | "},
						element("\\hline"),
						element("\\row",
							element("\\cell", par(element("\\bf", text("Group")), text(" "))),
							element("\\cell", par(text("\n    "), element("\\bf", text("Add. constraints")), text(" "))),
							element("\\cell", par(text("\n    "), element("\\bf", text("Points")), text(" "))),
							element("\\cell", par(text("\n    "), element("\\bf", text("Req. groups")), text(" "))),
						),
						element("\\hline"),
						element("\\row",
							element("\\cell", par(element("$", text("1")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("b = a + 1")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("30")), text(" "))),
							element("\\cell", par(text(" — "))),
						),
						element("\\hline"),
						element("\\row",
							element("\\cell", par(element("$", text("2")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("n \\le 1\\,000")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("10")), text(" "))),
							element("\\cell", par(text(" examples "))),
						),
						element("\\hline"),
						element("\\row",
							element("\\cell", par(element("$", text("3")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("n \\le 10^7")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("20")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("2")), text(" "))),
						),
						element("\\hline"),
						element("\\row",
							element("\\cell", par(element("$", text("4")), text(" "))),
							element("\\cell", par(text(" — "))),
							element("\\cell", par(text(" "), element("$", text("40")), text(" "))),
							element("\\cell", par(text(" "), element("$", text("1")), text(", "), element("$", text("3")), text(" "))),
						),
						element("\\hline"),
					),
					par(text("\n")),
				),
			),
		},
		{
			name:  "cf34",
			input: "\\begin{center}\n  \\begin{tabular}{cc}\n    \\includegraphics{eolymp.png} &\n    \\includegraphics{eolymp.png}\n  \\end{tabular}\n  \\small{Images side by side example.}\n\\end{center}",
			output: doc(
				element("center",
					par(text("\n  ")),
					elementp("tabular", map[string]string{"colspec": "cc"},
						element("\\row",
							element("\\cell", par(text("\n    ")), elementp("\\includegraphics", map[string]string{"src": "eolymp.png"}), par(text(" "))),
							element("\\cell", par(text("\n    ")), elementp("\\includegraphics", map[string]string{"src": "eolymp.png"}), par(text("\n  "))),
						),
					),
					par(text("\n  "), element("\\small", text("Images side by side example.")), text("\n")),
				),
			),
		},
		{
			name:  "cf37",
			input: "If you want to quote single character, use single quotes: `a'.\n\nIn some statements use <<these double quotes>>. As for the long dashes~--- use these like that.\n\nIn English statements use ``these double quotes''. As for the long dashes \"--- use these like that.",
			output: doc(
				par(text("If you want to quote single character, use single quotes: 'a'.\n")),
				par(text("In some statements use «these double quotes». As for the long dashes — use these like that.\n")),
				par(text("In English statements use \"these double quotes\". As for the long dashes \"— use these like that.")),
			),
		},
		{
			name:  "cf38",
			input: "\\epigraph{\\it{Some inspirational citation...}}{--- Author of citation, \\it{Source}}\nLegend starts here...",
			output: doc(
				element("\\epigraph",
					element("\\epigraph:text", element("\\it", text("Some inspirational citation..."))),
					element("\\epigraph:source", text("— Author of citation, "), element("\\it", text("Source"))),
				),
				par(text("\nLegend starts here...")),
			),
		},
		{
			name:  "problem environment",
			input: "\\begin{problem}{Шахівниця}{standard input}{standard output}{1 second}{256 megabytes} \n \nДано шахівницю $8\\times 8$. \\end{problem}",
			output: doc(
				elementp("problem", map[string]string{"title": "Шахівниця", "input": "standard input", "output": "standard output", "time_limit": "1 second", "memory_limit": "256 megabytes"},
					par(text(" \n")),
					par(text("Дано шахівницю "), element("$", text("8\\times 8")), text(". ")),
				),
			),
		},
		{
			name:   "example environment",
			input:  "\\begin{example}\nfoobar\\end{example}",
			output: doc(element("example", par(text("\nfoobar")))),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			parser := latex.NewParser(strings.NewReader(tc.input))

			got, err := parser.Parse()
			if err != nil {
				t.Fatalf("Unable to parse document: %v", err)
			}

			want := tc.output

			if !reflect.DeepEqual(want, got) {
				w, _ := json.MarshalIndent(want, "  ", "  ")
				g, _ := json.MarshalIndent(got, "  ", "  ")

				t.Errorf("Tree does not match:\nWANT:\n  %s\nGOT:\n  %s\n", w, g)
			}
		})
	}
}
