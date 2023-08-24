package latex

func symbol(a string) string {
	switch a {
	case "---":
		return "—"
	case "--":
		return "–"
	//case "<":
	//	return "‹"
	case "<<":
		return "«"
	//case ">":
	//	return "›"
	case ">>":
		return "»"
	case "''", "``":
		return "\""
	case "'", "`":
		return "'"
	case "&":
		return ""
	case "~":
		return string([]rune{0x00A0})
	default:
		return a
	}
}
