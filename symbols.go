package latex

func symbol(a string) string {
	switch a {
	case "---":
		return "—"
	case "--":
		return "–"
	case "<":
		return "‹"
	case "<<":
		return "«"
	case ">":
		return "›"
	case ">>":
		return "»"
	case "''", "``":
		return "\""
	case "'", "`":
		return "'"
	case "&":
		return ""
	case "~":
		return " "
	default:
		return a
	}
}
