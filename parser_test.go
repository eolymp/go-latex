package latex_test

import (
	"github.com/eolymp/go-latex"
	"github.com/google/go-cmp/cmp"

	"strings"
	"testing"
)

var nbsp = string([]rune{0x00A0})

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
		}, {
			name:  "media",
			input: "This is a video \\includemedia[scale=1.5]{eolymp.mov} some text after...",
			output: doc(
				par(text("This is a video ")),
				elementp("\\includemedia", map[string]string{
					"options": "scale=1.5",
					"src":     "eolymp.mov",
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
			name:  "lstlisting with language",
			input: "Some C++ source code (auto-detecting and highlighting):\n\\begin{lstlisting}[language=C++]\n#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n\\end{lstlisting}",
			output: doc(
				par(text("Some C++ source code (auto-detecting and highlighting):\n")),
				elementp("lstlisting", map[string]string{"options": "language=C++"}, text("#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n")),
			),
		},
		{
			name:  "lstlisting with whitespace prefix",
			input: "\\begin{lstlisting}[language=C++]\n    int a, b;\n    std::cin >> a >> b;\n\\end{lstlisting}",
			output: doc(
				elementp("lstlisting", map[string]string{"options": "language=C++"}, text("    int a, b;\n    std::cin >> a >> b;\n")),
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
				par(text(" ")),
				element("\\\\"),
				par(
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
				par(text(" ")),
				element("\\\\"),
				par(
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
				par(text(" ")),
				element("\\\\"),
				par(element("\\small", text("Centered image with width specified (180px).")), text("\n")),
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
			name:  "cf35",
			input: "Advanced scoring table example (colspan and rowspan):\n\\begin{tabular}{|c|c|c|c|c|}\n   \\hline\n    \\multirow{2}{*}{\\bf{Group}} &\n    \\multicolumn{2}{c|}{\\bf{Add. constraints}} &\n    \\multirow{2}{*}{\\bf{Points}} &\n    \\multirow{2}{*}{\\bf{Req. groups}} \\\\ \\cline{2-3}\n      & $n$ & $a_i$ & & \\\\ \\hline\n    $1$ & $n \\le 10$ & --- & $12$ & --- \\\\ \\hline\n    $2$ & $n \\le 500$ & $a_i \\le 100$ & $19$ & --- \\\\ \\hline\\end{tabular}",
			output: doc(
				par(text("Advanced scoring table example (colspan and rowspan):\n")),
				elementp("tabular", map[string]string{"colspec": "|c|c|c|c|c|"},
					element("\\hline"),
					element("\\row",
						elementp("\\cell", map[string]string{"rowspan": "2", "width": "*"}, par(element("\\bf", text("Group")))),
						elementp("\\cell", map[string]string{"colspan": "2", "align": "c|"}, par(element("\\bf", text("Add. constraints")))),
						elementp("\\cell", map[string]string{"rowspan": "2", "width": "*"}, par(element("\\bf", text("Points")))),
						elementp("\\cell", map[string]string{"rowspan": "2", "width": "*"}, par(element("\\bf", text("Req. groups")))),
					),
					elementp("\\cline", map[string]string{"range": "2-3"}),
					element("\\row",
						element("\\cell", par(text("\n      "))),
						element("\\cell", par(text(" "), element("$", text("n")), text(" "))),
						element("\\cell", par(text(" "), element("$", text("a_i")), text(" "))),
						element("\\cell", par(text(" "))),
						element("\\cell", par(text(" "))),
					),
					element("\\hline"),
					element("\\row",
						element("\\cell", par(element("$", text("1")), text(" "))),
						element("\\cell", par(text(" "), element("$", text("n \\le 10")), text(" "))),
						element("\\cell", par(text(" — "))),
						element("\\cell", par(text(" "), element("$", text("12")), text(" "))),
						element("\\cell", par(text(" — "))),
					),
					element("\\hline"),
					element("\\row",
						element("\\cell", par(element("$", text("2")), text(" "))),
						element("\\cell", par(text(" "), element("$", text("n \\le 500")), text(" "))),
						element("\\cell", par(text(" "), element("$", text("a_i \\le 100")), text(" "))),
						element("\\cell", par(text(" "), element("$", text("19")), text(" "))),
						element("\\cell", par(text(" — "))),
					),
					element("\\hline"),
				),
			),
		},
		{
			name:  "cf37",
			input: "If you want to quote single character, use single quotes: `a'.\n\nIn some statements use <<these double quotes>>. As for the long dashes~--- use these like that.\n\nIn English statements use ``these double quotes''. As for the long dashes~--- use these like that.",
			output: doc(
				par(text("If you want to quote single character, use single quotes: 'a'.\n")),
				par(text("In some statements use «these double quotes». As for the long dashes"+nbsp+"— use these like that.\n")),
				par(text("In English statements use \"these double quotes\". As for the long dashes"+nbsp+"— use these like that.")),
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
			input: "\\begin{problem}{Шахівниця}{standard render}{standard document}{1 second}{256 megabytes} \n \nДано шахівницю $8\\times 8$. \\end{problem}",
			output: doc(
				elementp("problem", map[string]string{"title": "Шахівниця", "input": "standard render", "output": "standard document", "time_limit": "1 second", "memory_limit": "256 megabytes"},
					par(text(" \n")),
					par(text("Дано шахівницю "), element("$", text("8\\times 8")), text(". ")),
				),
			),
		},
		{
			name:  "tutorial environment",
			input: "\\begin{tutorial}{Шахівниця}how to solve...\\end{tutorial}",
			output: doc(
				elementp("tutorial", map[string]string{"title": "Шахівниця"},
					par(text("how to solve...")),
				),
			),
		},
		{
			name:   "example environment",
			input:  "\\begin{example}\nfoobar\\end{example}",
			output: doc(element("example", par(text("\nfoobar")))),
		},
		{
			name:   "whitespace between parameterString",
			input:  "\\includegraphics[width=5cm, height=5cm] {xx.png}",
			output: doc(elementp("\\includegraphics", map[string]string{"options": "width=5cm, height=5cm", "src": "xx.png"})),
		},
		{
			name:  "p10675",
			input: "\\begin{center}\n{\\includegraphics{https://static.eolymp.com/content/2c/2cb0e289dc31d026e2c5481852803fe3a0b8c38b.png}}\\end{center}",
			output: doc(element("center",
				par(text("\n")),
				elementp("\\includegraphics", map[string]string{"src": "https://static.eolymp.com/content/2c/2cb0e289dc31d026e2c5481852803fe3a0b8c38b.png"}),
			)),
		},
		{
			name:  "unbound empty group",
			input: "foo {} baz",
			output: doc(par(
				text("foo  baz"),
			)),
		},
		{
			name:  "unbound text group",
			input: "foo {bar \\textit{bug}} baz",
			output: doc(par(
				text("foo "),
				element("{}", text("bar "), element("\\textit", text("bug"))),
				text(" baz"),
			)),
		},
		{
			name:  "unbound block group",
			input: "foo {one\n\n two} baz",
			output: doc(
				par(text("foo ")),
				element("{}", par(text("one\n")), par(text(" two"))),
				par(text(" baz")),
			),
		},
		{
			name:  "p12360",
			input: "\\begin{wrapfigure}{r}{0.30}\n\\vspace{-20pt}\n  \\begin{center}\n    \\includegraphics[width=0.30]{pic.jpg}\n  \\end{center}\n  \\vspace{-20pt}\n  \\vspace{1pt}\n\\end{wrapfigure}\n",
			output: doc(
				elementp("wrapfigure", map[string]string{"position": "r", "width": "0.30"},
					par(text("\n")),
					elementp("\\vspace", map[string]string{"height": "-20pt"}),
					par(text("  ")),
					element("center",
						par(text("\n    ")),
						elementp("\\includegraphics", map[string]string{"options": "width=0.30", "src": "pic.jpg"}),
						par(text("\n  ")),
					),
					par(text("\n  ")),
					elementp("\\vspace", map[string]string{"height": "-20pt"}),
					par(text("\n  ")),
					elementp("\\vspace", map[string]string{"height": "1pt"}),
					par(text("\n")),
				),
				par(text("\n")),
			),
		},
		{
			name:   "p12587",
			input:  "\\includegraphics{https://foo.com/www.bar.com/wp-content/uploads/2021/02/4cbe8d_f1ed2800a49649848102c68fc5a66e53mv2.gif?fit=476%2C280&ssl=1}",
			output: doc(elementp("\\includegraphics", map[string]string{"src": "https://foo.com/www.bar.com/wp-content/uploads/2021/02/4cbe8d_f1ed2800a49649848102c68fc5a66e53mv2.gif?fit=476%2C280&ssl=1"})),
		},
		{
			name:  "p12854",
			input: "\\epigraph{Hello, and again, welcome to the Aperture Science Enrichment Center.}",
			output: doc(element("\\epigraph",
				element("\\epigraph:text", text("Hello, and again, welcome to the Aperture Science Enrichment Center.")),
				element("\\epigraph:source"),
			)),
		},
		{
			name:  "command in group",
			input: "foo {\\it Hello, and again, welcome to the Aperture Science Enrichment Center.} bar",
			output: doc(par(
				text("foo "),
				element("\\it", text("Hello, and again, welcome to the Aperture Science Enrichment Center.")),
				text(" bar"),
			)),
		},
		{
			name:  "user mention",
			input: "i would like \\user{arsijo} to be a judge of this",
			output: doc(par(
				text("i would like "),
				elementp("\\user", map[string]string{"nickname": "arsijo"}),
				text(" to be a judge of this"),
			)),
		},
		{
			name:   "verbatim parameter with {}",
			input:  "\\exmp{\\{[]\\}}{OK}",
			output: doc(elementp("\\exmp", map[string]string{"input": "{[]}", "output": "OK"})),
		},
		{
			name:  "custom environment with parameter",
			input: "\\begin{grid}[columns=6]\n  This content is in the block.\n\n  $abacaba$\n\\end{grid}",
			output: doc(elementp("grid",
				map[string]string{"options": "columns=6"},
				par(text("\n  This content is in the block.\n")),
				par(text("  "), element("$", text("abacaba")), text("\n")),
			)),
		},
		{
			name:  "heading",
			input: "\\heading[3]{Level three heading}",
			output: doc(par(elementp("\\heading",
				map[string]string{"level": "3"},
				text("Level three heading"),
			))),
		},
		{
			name:  "tabs",
			input: "\\begin{tabs}\n  \\item{Tab 1} This is the first item;\n  \\item{Tab 2} This is the second item.\n\\end{tabs}",
			output: doc(
				element("tabs",
					elementp("\\item", map[string]string{"title": "Tab 1"}, par(text(" This is the first item;\n  "))),
					elementp("\\item", map[string]string{"title": "Tab 2"}, par(text(" This is the second item.\n"))),
				),
			),
		},
		{
			name:  "admonition with cyrillic letters",
			input: "\\begin{admonition}[type=note, title=\"Привіт 👋\"]Як справи? ⁉️\\end{admonition}",
			output: doc(elementp("admonition",
				map[string]string{"options": "type=note, title=\"Привіт 👋\""},
				par(text("Як справи? ⁉️")),
			)),
		},
		{
			name:  "divider",
			input: "123\\hline456\\hrule789",
			output: doc(
				par(text("123")),
				element("\\hline"),
				par(text("456")),
				element("\\hrule"),
				par(text("789")),
			),
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

			if !cmp.Equal(want, got) {
				t.Errorf("Tree does not match:\n%s\n", cmp.Diff(want, got))
			}
		})
	}
}
