package latex

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func Render(w io.Writer, node *Node) error {
	return render(w, node)
}

func render(w io.Writer, node *Node) error {
	switch node.Kind {
	case DocumentKind:
		return renderChildren(w, node)
	case TextKind:
		return renderText(w, node)
	case ElementKind:
		return renderElement(w, node)
	default:
		return nil
	}
}

func renderText(w io.Writer, node *Node) error {
	value := node.Data
	for f, t := range specials {
		if t == "" {
			continue
		}

		value = strings.ReplaceAll(value, f, t)
	}

	//for f, t := range replacements {
	//	if t == "" || t == " " {
	//		continue
	//	}
	//
	//	value = strings.ReplaceAll(value, t, f)
	//}

	_, err := fmt.Fprint(w, value)
	return err
}

func renderVerbatim(w io.Writer, node *Node) error {
	if node.Kind == TextKind {
		if _, err := fmt.Fprint(w, node.Data); err != nil {
			return err
		}
	}

	for _, child := range node.Children {
		if err := renderVerbatim(w, child); err != nil {
			return err
		}
	}

	return nil
}

func renderChildren(w io.Writer, node *Node) error {
	for _, child := range node.Children {
		if err := render(w, child); err != nil {
			return err
		}
	}

	return nil
}

func renderChildrenAndWrap(node *Node, w io.Writer, prefix, suffix string) error {
	if _, err := fmt.Fprint(w, prefix); err != nil {
		return err
	}

	if err := renderChildren(w, node); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, suffix); err != nil {
		return err
	}

	return nil
}

func renderVerbatimAndWrap(node *Node, w io.Writer, prefix, suffix string) error {
	if _, err := fmt.Fprint(w, prefix); err != nil {
		return err
	}

	if err := renderVerbatim(w, node); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, suffix); err != nil {
		return err
	}

	return nil
}

func renderElement(w io.Writer, node *Node) error {
	switch node.Data {
	case "\\par":
		return renderChildrenAndWrap(node, w, "", "\n\n")
	case "\\\\", "\\\\*", "\\newline":
		_, err := fmt.Fprint(w, node.Data+"\n")
		return err

	case "\\InputFile", "\\InputData", "\\OutputFile", "\\Note", "\\Scoring", "\\Interaction", "\\Example", "\\Examples":
		_, err := fmt.Fprint(w, node.Data, "\n\n")
		return err
	case "\\dots", "\\ldots", "\\cdots", "\\vdots", "\\ddots", "\\hskip", "\\vskip", "\\hline", "\\cline", "\\multicolumn", "\\vspace", "\\hspace":
		_, err := fmt.Fprint(w, node.Data)
		return err
	case "\\epigraph":
		return nil
	case "\\epigraph:text", "\\epigraph:source":
		return nil
	case "\\item":
		return renderChildrenAndWrap(node, w, "\\item ", "")
	case "\\verb", "\\verb*":
		delimiter := node.Parameters["delimiter"]
		if delimiter == "" {
			delimiter = "|"
		}

		return renderVerbatimAndWrap(node, w, node.Data+delimiter, delimiter)
	case "verbatim":
		return renderVerbatimAndWrap(node, w, "\\begin{verbatim}\n", "\\end{verbatim}")
	case "lstlisting":
		params := ""
		if v := node.Parameters["options"]; v != "" {
			params = "[" + v + "]"
		}

		return renderVerbatimAndWrap(node, w, "\\begin{verbatim}"+params+"\n", "\\end{verbatim}")
	case "tabular":
		colspec := ""
		if v := node.Parameters["colspec"]; v != "" {
			colspec = "{" + v + "}"
		}

		var rows []string
		for index, child := range node.Children {
			if child.Kind == ElementKind && child.Data == "\\hline" {
				rows = append(rows, "\\hline")
				continue
			}

			buffer := bytes.NewBuffer(nil)
			if err := render(buffer, child); err != nil {
				return err
			}

			suffix := " \\\\"
			if index == len(node.Children)-1 {
				suffix = ""
			}

			rows = append(rows, strings.TrimSpace(buffer.String())+suffix)
		}

		_, err := fmt.Fprint(w, "\\begin{tabular}"+colspec+"\n", strings.Join(rows, "\n"), "\n\\end{tabular}")
		return err
	case "itemize", "enumerate", "center", "example":
		return renderChildrenAndWrap(node, w, "\\begin{"+node.Data+"}\n", "\\end{"+node.Data+"}")
	case "{}":
		return renderChildren(w, node)
	case "\\row":
		var cells []string
		for _, child := range node.Children {
			buffer := bytes.NewBuffer(nil)
			if err := render(buffer, child); err != nil {
				return err
			}

			cells = append(cells, strings.TrimSpace(buffer.String()))
		}

		_, err := fmt.Fprint(w, strings.Join(cells, " & "))
		return err
	case "\\cell":
		return renderChildren(w, node)
	case "$":
		return renderVerbatimAndWrap(node, w, "$", "$")
	case "$$":
		return renderVerbatimAndWrap(node, w, "$$", "$$")
	case "%", "comment":
		return nil
	case "\\symbol":
		return nil
	case "\\underline", "\\emph", "\\sout", "\\textmd", "\\textbf", "\\textup", "\\textit", "\\textsl", "\\textsc", "\\textsf", "\\textrm", "\\bf", "\\it", "\\t", "\\tt", "\\texttt", "\\tiny", "\\scriptsize", "\\small", "\\normalsize", "\\large", "\\Large", "\\LARGE", "\\huge", "\\Huge", "\\section", "\\subsection", "\\subsubsection", "\\bfseries", "\\itshape":
		if _, err := fmt.Fprint(w, node.Data+"{"); err != nil {
			return err
		}

		if err := renderChildren(w, node); err != nil {
			return err
		}

		_, err := fmt.Fprint(w, "}")
		return err

	case "\\includegraphics":
		src, _ := node.Parameters["src"]
		params := ""

		if opts, ok := node.Parameters["options"]; ok {
			params = "[" + opts + "]"
		}

		_, err := fmt.Fprint(w, "\\includegraphics", params, "{", src, "}")
		return err

	case "\\url":
		_, err := fmt.Fprint(w, "\\url{", node.Parameters["href"], "}")
		return err
	case "\\href":
		return renderChildrenAndWrap(node, w, "\\href{"+node.Parameters["href"]+"}{", "}")
	case "\\def":
		return nil
	case "\\exmp":
		return nil
	case "\\exmpfile":
		return nil
	case "\\user":
		_, err := fmt.Fprint(w, "\\user{", node.Parameters["nickname"], "}")
		return err

	default:
		return nil
	}
}
