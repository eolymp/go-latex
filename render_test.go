package latex_test

import (
	"bytes"
	"github.com/eolymp/go-latex"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
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
		name     string
		render   string
		document *latex.Node
	}{
		{
			name:     "simple paragraph",
			render:   "one two\nthree",
			document: doc(par(text("one two\nthree"))),
		},
		{
			name:   "two paragraphs separated by new lines",
			render: "one\ntwo\n\n\nthree\n\n\nfour",
			document: doc(
				par(text("one\ntwo\n")),
				par(text("three\n")),
				par(text("four")),
			),
		},
		{
			name:   "simple formatting",
			render: "odd \\textbf{foo bar} baz",
			document: doc(par(
				text("odd "),
				element("\\textbf", text("foo bar")),
				text(" baz"),
			)),
		},
		{
			name:   "nested formatting",
			render: "odd \\textbf{foo \\textit{bar}} baz",
			document: doc(par(
				text("odd "),
				element("\\textbf", text("foo "), element("\\textit", text("bar"))),
				text(" baz"),
			)),
		},
		{
			name:   "inline math",
			render: "$\\alpha + \\beta$",
			document: doc(par(
				element("$", text("\\alpha + \\beta")),
			)),
		},
		{
			name:     "block math",
			render:   "$$\\alpha + \\beta$$",
			document: doc(element("$$", text("\\alpha + \\beta"))),
		},
		{
			name:   "math",
			render: "foo $a_i^2 + b_i^2 \\le a_{i+1}^2$ bar",
			document: doc(par(
				text("foo "),
				element("$", text("a_i^2 + b_i^2 \\le a_{i+1}^2")),
				text(" bar"),
			)),
		},
		{
			name:   "math block",
			render: "foo \n\n$$a_i^2 + b_i^2 \\le a_{i+1}^2$$ bar",
			document: doc(
				par(text("foo ")),
				element("$$", text("a_i^2 + b_i^2 \\le a_{i+1}^2")),
				par(text(" bar")),
			),
		},
		{
			name:   "image",
			render: "This is an image \n\n\\includegraphics[scale=1.5]{eolymp.png} some text after...",
			document: doc(
				par(text("This is an image ")),
				elementp("\\includegraphics", map[string]string{
					"options": "scale=1.5",
					"src":     "eolymp.png",
				}),
				par(text(" some text after...")),
			),
		},
		{
			name:   "verb command",
			render: "The \\verb|\\ldots| command \\ldots",
			document: doc(par(
				text("The "),
				element("\\verb", text("\\ldots")),
				text(" command "),
				element("\\ldots"),
			)),
		},
		{
			name:   "verbatim environment",
			render: "the \n\n\\begin{verbatim}\n10 PRINT \"HELLO WORLD \";\n20 GOTO 10\n\\end{verbatim} code",
			document: doc(
				par(text("the ")),
				element("verbatim", text("10 PRINT \"HELLO WORLD \";\n20 GOTO 10\n")),
				par(text(" code")),
			),
		},
		{
			name:     "verb command with star",
			render:   "\\verb*|like   this :-) |",
			document: doc(par(element("\\verb*", text("like   this :-) ")))),
		},
		{
			name:   "cf1",
			render: "These are inline formulas: $x$, $a_i^2 + b_i^2 \\le a_{i+1}^2$. Afterwards...",
			document: doc(par(
				text("These are inline formulas: "),
				element("$", text("x")),
				text(", "),
				element("$", text("a_i^2 + b_i^2 \\le a_{i+1}^2")),
				text(". Afterwards..."),
			)),
		},
		{
			name:   "cf2",
			render: "These are centered formulas: \n\n$$x,$$ \n\n$$a_i^2 + b_i^2 \\le a_{i+1}^2.$$ Afterwards...",
			document: doc(
				par(text("These are centered formulas: ")),
				element("$$", text("x,")),
				par(text(" ")),
				element("$$", text("a_i^2 + b_i^2 \\le a_{i+1}^2.")),
				par(text(" Afterwards...")),
			),
		},
		{
			name:   "cf3",
			render: "Some complex formula: \n\n$$P(|S - E[S]| \\ge t) \\le 2 \\exp \\left( -\\frac{2 t^2 n^2}{\\sum_{i = 1}^n (b_i - a_i)^2} \\right).$$",
			document: doc(
				par(text("Some complex formula: ")),
				element("$$", text("P(|S - E[S]| \\ge t) \\le 2 \\exp \\left( -\\frac{2 t^2 n^2}{\\sum_{i = 1}^n (b_i - a_i)^2} \\right).")),
			),
		},
		{
			name:     "cf4",
			render:   "First paragraph.\nStill first paragraph.",
			document: doc(par(text("First paragraph.\nStill first paragraph."))),
		},
		{
			name:   "cf5",
			render: "First paragraph.\n\n\nSecond paragraph.",
			document: doc(
				par(text("First paragraph.\n")),
				par(text("Second paragraph.")),
			),
		},
		{
			name:   "cf6",
			render: "\\bf{This text is bold.}or\n\\textbf{This text is bold.}",
			document: doc(par(
				element("\\bf", text("This text is bold.")),
				text("or\n"),
				element("\\textbf", text("This text is bold.")),
			)),
		},
		{
			name:   "cf7",
			render: "\\it{This text is italic.}or\n\\textit{This text is italic.}",
			document: doc(par(
				element("\\it", text("This text is italic.")),
				text("or\n"),
				element("\\textit", text("This text is italic.")),
			)),
		},
		{
			name:   "cf8",
			render: "\\t{This text is monospaced.}or\n\\tt{This text is monospaced.}or\n\\texttt{This text is monospaced.}",
			document: doc(par(
				element("\\t", text("This text is monospaced.")),
				text("or\n"),
				element("\\tt", text("This text is monospaced.")),
				text("or\n"),
				element("\\texttt", text("This text is monospaced.")),
			)),
		},
		{
			name:   "cf9",
			render: "\\emph{This text is underlined.}or\n\\underline{This text is underlined.}",
			document: doc(par(
				element("\\emph", text("This text is underlined.")),
				text("or\n"),
				element("\\underline", text("This text is underlined.")),
			)),
		},
		{
			name:     "cf10",
			render:   "\\sout{This text is struck out.}",
			document: doc(par(element("\\sout", text("This text is struck out.")))),
		},
		{
			name:     "cf11",
			render:   "\\textsc{This text is capitalized.}",
			document: doc(par(element("\\textsc", text("This text is capitalized.")))),
		},
		{
			name:     "cf12",
			render:   "\\tiny{This text is 70\\% of normal size.}",
			document: doc(par(element("\\tiny", text("This text is 70% of normal size.")))),
		},
		{
			name:     "cf13",
			render:   "\\scriptsize{This text is 75\\% of normal size.}",
			document: doc(par(element("\\scriptsize", text("This text is 75% of normal size.")))),
		},
		{
			name:     "cf14",
			render:   "\\small{This text is 85\\% of normal size.}",
			document: doc(par(element("\\small", text("This text is 85% of normal size.")))),
		},
		{
			name:   "cf15",
			render: "\\normalsize{This text is 100\\% of normal size.}\nor just\nThis text is 100\\% of normal size.",
			document: doc(par(
				element("\\normalsize", text("This text is 100% of normal size.")),
				text("\nor just\nThis text is 100% of normal size."),
			)),
		},
		{
			name:   "cf16..20",
			render: "\\large{This text is 115\\% of normal size.}\\Large{This text is 130\\% of normal size.}\\LARGE{This text is 145\\% of normal size.}\\huge{This text is 175\\% of normal size.}\\Huge{This text is 200\\% of normal size.}",
			document: doc(par(
				element("\\large", text("This text is 115% of normal size.")),
				element("\\Large", text("This text is 130% of normal size.")),
				element("\\LARGE", text("This text is 145% of normal size.")),
				element("\\huge", text("This text is 175% of normal size.")),
				element("\\Huge", text("This text is 200% of normal size.")),
			)),
		},
		{
			name:   "cf21",
			render: "This is the unordered list:\n\n\n\\begin{itemize}\n\\item This is the first item;\n\n\n\\item This is the second item.\n\n\n\\end{itemize}",
			document: doc(
				par(text("This is the unordered list:\n")),
				element("itemize",
					element("\\item", par(text("This is the first item;\n"))),
					element("\\item", par(text("This is the second item.\n"))),
				),
			),
		},
		{
			name:   "cf22",
			render: "This is the ordered list:\n\n\n\\begin{enumerate}\n\\item This is the first item;\n  \n\n\\item This is the second item.\n\n\n\\end{enumerate}",
			document: doc(
				par(text("This is the ordered list:\n")),
				element("enumerate",
					element("\\item", par(text("This is the first item;\n  "))),
					element("\\item", par(text("This is the second item.\n"))),
				),
			),
		},
		{
			name:   "cf23",
			render: "Some C++ source code (auto-detecting and highlighting):\n\n\n\\begin{verbatim}\n#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n\\end{verbatim}",
			document: doc(
				par(text("Some C++ source code (auto-detecting and highlighting):\n")),
				element("lstlisting", text("#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n")),
			),
		},
		{
			name:   "lstlisting with language",
			render: "Some C++ source code (auto-detecting and highlighting):\n\n\n\\begin{verbatim}[language=C++]\n#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n\\end{verbatim}",
			document: doc(
				par(text("Some C++ source code (auto-detecting and highlighting):\n")),
				elementp("lstlisting", map[string]string{"options": "language=C++"}, text("#include <iostream>\nint main() {\n    int a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << std::endl;\n}\n")),
			),
		},
		{
			name:   "cf24",
			render: "Link to website:\n\\url{https://eolymp.com/}.",
			document: doc(par(
				text("Link to website:\n"),
				elementp("\\url", map[string]string{"href": "https://eolymp.com/"}),
				text("."),
			)),
		},
		{
			name:   "cf25",
			render: "Link to website with caption:\n\\href{https://eolymp.com/}{Eolymp}.",
			document: doc(par(
				text("Link to website with caption:\n"),
				elementp("\\href", map[string]string{"href": "https://eolymp.com/"}, text("Eolymp")),
				text("."),
			)),
		},
		{
			name:   "cf26",
			render: "\\begin{center}\n\n  This content is centered.\n\n\n  $abacaba$\n\n\n\\end{center}",
			document: doc(element("center",
				par(text("\n  This content is centered.\n")),
				par(text("  "), element("$", text("abacaba")), text("\n")),
			)),
		},
		{
			name:   "cf27",
			render: "Unscaled image:\n\n\n\\includegraphics{eolymp.png}",
			document: doc(
				par(text("Unscaled image:\n")),
				elementp("\\includegraphics", map[string]string{"src": "eolymp.png"}),
			),
		},
		{
			name:   "cf28",
			render: "\\begin{center}\n\n  \n\n\\includegraphics{eolymp.png} \n\n\\\\\n\\small{Centered unscaled image.}\n\n\n\\end{center}",
			document: doc(element("center",
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
			name:   "cf29",
			render: "\\begin{center}\n\n  \n\n\\includegraphics[scale=1.5]{eolymp.png} \n\n\\\\\n\\small{Centered scaled image.}\n\n\n\\end{center}",
			document: doc(element("center",
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
			name:   "cf31",
			render: "Simple table without borders:\n\n\n\\begin{tabular}{ll}\nFirst & Second \\\\\nThird & Fourth\n\\end{tabular}",
			document: doc(
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
			name:   "cf32",
			render: "More complex table with borders:\n\n\n\\begin{tabular}{|l|c|r|}\n\\hline\nLeft aligned column & Centered column & Right aligned column \\\\\n\\hline\nText & Text & Text \\\\\n\\hline\n\\end{tabular}",
			document: doc(
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
			name:   "cf33",
			render: "Scoring table example:\n\n\n\\begin{center}\n\n  \n\n\\begin{tabular}{ | c | c | c | c | }\n\\hline\n\\bf{Group} & \\bf{Add. constraints} & \\bf{Points} & \\bf{Req. groups} \\\\\n\\hline\n$1$ & $b = a + 1$ & $30$ & --- \\\\\n\\hline\n$2$ & $n \\le 1\\,000$ & $10$ & examples \\\\\n\\hline\n$3$ & $n \\le 10^7$ & $20$ & $2$ \\\\\n\\hline\n$4$ & --- & $40$ & $1$, $3$ \\\\\n\\hline\n\\end{tabular}\n\n\n\\end{center}",
			document: doc(
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
			name:   "cf34",
			render: "\\begin{center}\n\n  \n\n\\begin{tabular}{cc}\n\\includegraphics{eolymp.png} & \\includegraphics{eolymp.png}\n\\end{tabular}\n  \\small{Images side by side example.}\n\n\n\\end{center}",
			document: doc(
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
			name:   "cf37",
			render: "If you want to quote single character, use single quotes: a.\n\n\nIn some statements use these double quotes. As for the long dashes~--- use these like that.\n\n\nIn English statements use these double quotes. As for the long dashes~--- use these like that.",
			document: doc(
				par(text("If you want to quote single character, use single quotes: a.\n")),
				par(text("In some statements use these double quotes. As for the long dashes"+nbsp+"— use these like that.\n")),
				par(text("In English statements use these double quotes. As for the long dashes"+nbsp+"— use these like that.")),
			),
		},
		//{
		//	name:   "cf38",
		//	render: "\\epigraph{\\it{Some inspirational citation...}}{--- Author of citation, \\it{Source}}\nLegend starts here...",
		//	document: doc(
		//		element("\\epigraph",
		//			element("\\epigraph:text", element("\\it", text("Some inspirational citation..."))),
		//			element("\\epigraph:source", text("— Author of citation, "), element("\\it", text("Source"))),
		//		),
		//		par(text("\nLegend starts here...")),
		//	),
		//},
		//{
		//	name:   "problem environment",
		//	render: "\\begin{problem}{Шахівниця}{standard render}{standard document}{1 second}{256 megabytes} \n \nДано шахівницю $8\\times 8$. \\end{problem}",
		//	document: doc(
		//		elementp("problem", map[string]string{"title": "Шахівниця", "render": "standard render", "document": "standard document", "time_limit": "1 second", "memory_limit": "256 megabytes"},
		//			par(text(" \n")),
		//			par(text("Дано шахівницю "), element("$", text("8\\times 8")), text(". ")),
		//		),
		//	),
		//},
		//{
		//	name:   "tutorial environment",
		//	render: "\\begin{tutorial}{Шахівниця}how to solve...\\end{tutorial}",
		//	document: doc(
		//		elementp("tutorial", map[string]string{"title": "Шахівниця"},
		//			par(text("how to solve...")),
		//		),
		//	),
		//},
		{
			name:     "example environment",
			render:   "\\begin{example}\n\nfoobar\n\n\\end{example}",
			document: doc(element("example", par(text("\nfoobar")))),
		},
		{
			name:   "p10675",
			render: "\\begin{center}\n\n\n\n\\includegraphics{https://static.eolymp.com/content/2c/2cb0e289dc31d026e2c5481852803fe3a0b8c38b.png}\\end{center}",
			document: doc(element("center",
				par(text("\n")),
				elementp("\\includegraphics", map[string]string{"src": "https://static.eolymp.com/content/2c/2cb0e289dc31d026e2c5481852803fe3a0b8c38b.png"}),
			)),
		},
		//{
		//	name:   "p12360",
		//	render: "\\begin{wrapfigure}{r}{0.30}\n\\vspace{-20pt}\n  \\begin{center}\n    \\includegraphics[width=0.30]{pic.jpg}\n  \\end{center}\n  \\vspace{-20pt}\n  \\vspace{1pt}\n\\end{wrapfigure}\n",
		//	document: doc(
		//		elementp("wrapfigure", map[string]string{"position": "r", "width": "0.30"},
		//			par(text("\n")),
		//			elementp("\\vspace", map[string]string{"height": "-20pt"}),
		//			par(text("  ")),
		//			element("center",
		//				par(text("\n    ")),
		//				elementp("\\includegraphics", map[string]string{"options": "width=0.30", "src": "pic.jpg"}),
		//				par(text("\n  ")),
		//			),
		//			par(text("\n  ")),
		//			elementp("\\vspace", map[string]string{"height": "-20pt"}),
		//			par(text("\n  ")),
		//			elementp("\\vspace", map[string]string{"height": "1pt"}),
		//			par(text("\n")),
		//		),
		//		par(text("\n")),
		//	),
		//},
		{
			name:     "p12587",
			render:   "\\includegraphics{https://foo.com/www.bar.com/wp-content/uploads/2021/02/4cbe8d_f1ed2800a49649848102c68fc5a66e53mv2.gif?fit=476%2C280&ssl=1}",
			document: doc(elementp("\\includegraphics", map[string]string{"src": "https://foo.com/www.bar.com/wp-content/uploads/2021/02/4cbe8d_f1ed2800a49649848102c68fc5a66e53mv2.gif?fit=476%2C280&ssl=1"})),
		},
		//{
		//	name:   "p12854",
		//	render: "\\epigraph{Hello, and again, welcome to the Aperture Science Enrichment Center.}",
		//	document: doc(element("\\epigraph",
		//		element("\\epigraph:text", text("Hello, and again, welcome to the Aperture Science Enrichment Center.")),
		//		element("\\epigraph:source"),
		//	)),
		//},
		//{
		//	name:   "command in group",
		//	render: "foo {\\it Hello, and again, welcome to the Aperture Science Enrichment Center.} bar",
		//	document: doc(par(
		//		text("foo "),
		//		element("\\it", text("Hello, and again, welcome to the Aperture Science Enrichment Center.")),
		//		text(" bar"),
		//	)),
		//},
		{
			name:   "user mention",
			render: "i would like \\user{arsijo} to be a judge of this",
			document: doc(par(
				text("i would like "),
				elementp("\\user", map[string]string{"nickname": "arsijo"}),
				text(" to be a judge of this"),
			)),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			buffer := bytes.NewBuffer(nil)

			err := latex.Render(buffer, tc.document)
			if err != nil {
				t.Fatal("unable to render:", err)
			}

			got := strings.TrimSuffix(buffer.String(), "\n\n")
			want := tc.render

			gsf := strings.ReplaceAll(strings.ReplaceAll(got, " ", ""), "\n", "")
			wsf := strings.ReplaceAll(strings.ReplaceAll(want, " ", ""), "\n", "")

			if got != want {
				if gsf == wsf {
					t.Errorf("Spaces are not properly placed:\nWANT:\n  %#v\nGOT:\n  %#v\n", want, got)
				} else {
					t.Errorf("Rendered latex does not match:\nWANT:\n  %#v\nGOT:\n  %#v\n", wsf, gsf)
				}

				t.Errorf("GOT:\n  %v\n", got)
			}
		})
	}
}
