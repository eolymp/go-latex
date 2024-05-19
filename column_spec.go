package latex

type ColumnSpec struct {
	BorderLeft  bool   // column should have left border
	BorderRight bool   // column should have right border
	Align       string // column alignment: c, l or r
}

// ColumnSpecs parses column spec in tabular environment
// todo: add support for repeated syntax *{x}{...}
// todo: if not support, at least correctly handle @{} and !{}
func ColumnSpecs(raw string) (spec []ColumnSpec) {
	raw = whitespaces.ReplaceAllString(raw, "") // remove all spaces since they don't have any meaning
	for pos, char := range raw {
		if char == '|' {
			continue
		}

		if char == 'c' || char == 'l' || char == 'r' {
			spec = append(spec, ColumnSpec{
				BorderLeft:  pos > 0 && raw[pos-1] == '|',
				BorderRight: pos < len(raw)-1 && raw[pos+1] == '|',
				Align:       string([]rune{char}),
			})
		}
	}

	return
}
