package latex

var replacements = map[string]string{
	"\\textwidth":            "",
	"\\space":                " ",
	"\\nobreakspace":         string([]rune{160}),
	"\\thinspace":            string([]rune{8201}),
	"\\enspace":              string([]rune{8194}),
	"\\enskip":               string([]rune{8194}),
	"\\quad":                 string([]rune{8195}),
	"\\qquad":                string([]rune{8195, 8195}),
	"\\textvisiblespace":     string([]rune{9251}),
	"\\textcompwordmark":     string([]rune{8204}),
	"\\textdollar":           "$",
	"\\slash":                "/",
	"\\textless":             "<",
	"\\textgreater":          ">",
	"\\textbackslash":        "\\",
	"\\textasciicircum":      "^",
	"\\textunderscore":       "_",
	"\\lbrack":               "[",
	"\\rbrack":               "]",
	"\\textbraceleft":        "\u007B",
	"\\textbraceright":       "\u007D",
	"\\textasciitilde":       "˜",
	"\\AA":                   "\u00C5",
	"\\aa":                   "\u00E5",
	"\\AE":                   "\u00C6",
	"\\ae":                   "\u00E6",
	"\\OE":                   "\u0152",
	"\\oe":                   "\u0153",
	"\\DH":                   "\u00D0",
	"\\dh":                   "\u00F0",
	"\\DJ":                   "\u0110",
	"\\dj":                   "\u0111",
	"\\NG":                   "\u014A",
	"\\ng":                   "\u014B",
	"\\TH":                   "\u00DE",
	"\\th":                   "\u00FE",
	"\\O":                    "\u00D8",
	"\\o":                    "\u00F8",
	"\\i":                    "\u0131",
	"\\j":                    "\u0237",
	"\\L":                    "\u0141",
	"\\l":                    "\u0142",
	"\\IJ":                   "\u0132",
	"\\ij":                   "\u0133",
	"\\SS":                   "\u1e9e",
	"\\ss":                   "\u00df",
	"\\textquotesingle":      "\"",
	"\\textquoteleft":        string([]rune{8216}),
	"\\lq":                   string([]rune{8216}),
	"\\textquoteright":       string([]rune{8217}),
	"\\rq":                   string([]rune{8217}),
	"\\textquotedbl":         string([]rune{34}),
	"\\textquotedblleft":     string([]rune{8220}),
	"\\textquotedblright":    string([]rune{8221}),
	"\\quotesinglbase":       string([]rune{8218}),
	"\\quotedblbase":         string([]rune{8222}),
	"\\guillemotleft":        string([]rune{171}),
	"\\guillemotright":       string([]rune{187}),
	"\\guilsinglleft":        string([]rune{8249}),
	"\\guilsinglright":       string([]rune{8250}),
	"\\textasciigrave":       "\u0060",
	"\\textgravedbl":         "\u02f5",
	"\\textasciidieresis":    string([]rune{168}),
	"\\textasciiacute":       string([]rune{180}),
	"\\textacutedbl":         string([]rune{733}),
	"\\textasciimacron":      string([]rune{175}),
	"\\textasciicaron":       string([]rune{711}),
	"\\textasciibreve":       string([]rune{728}),
	"\\texttildelow":         "\u02f7",
	"\\textendash":           string([]rune{8211}),
	"\\textemdash":           string([]rune{8212}),
	"\\textellipsis":         string([]rune{8230}),
	"\\dots":                 string([]rune{8230}),
	"\\ldots":                string([]rune{8230}),
	"\\textbullet":           string([]rune{8226}),
	"\\textopenbullet":       "\u25e6",
	"\\textperiodcentered":   string([]rune{183}),
	"\\textdagger":           string([]rune{8224}),
	"\\dag":                  string([]rune{8224}),
	"\\textdaggerdbl":        string([]rune{8225}),
	"\\ddag":                 string([]rune{8225}),
	"\\textexclamdown":       string([]rune{161}),
	"\\textquestiondown":     string([]rune{191}),
	"\\textinterrobang":      "\u203d",
	"\\textinterrobangdown":  "\u2e18",
	"\\textsection":          string([]rune{167}),
	"\\S":                    string([]rune{167}),
	"\\textparagraph":        string([]rune{182}),
	"\\P":                    string([]rune{182}),
	"\\textblank":            "\u2422",
	"\\textlquill":           "\u2045",
	"\\textrquill":           "\u2046",
	"\\textlangle":           "\u2329",
	"\\textrangle":           "\u232a",
	"\\textlbrackdbl":        "\u301a",
	"\\textrbrackdbl":        "\u301b",
	"\\textcopyright":        "©",
	"\\copyright":            "©",
	"\\textregistered":       string([]rune{174}),
	"\\textcircledP":         string([]rune{8471}),
	"\\textservicemark":      "\u2120",
	"\\texttrademark":        string([]rune{8482}),
	"\\textmarried":          "\u26ad",
	"\\textdivorced":         "\u26ae",
	"\\textordfeminine":      string([]rune{170}),
	"\\textordmasculine":     string([]rune{186}),
	"\\textdegree":           string([]rune{176}),
	"\\textmu":               string([]rune{181}),
	"\\textbar":              "\u007c",
	"\\textbardbl":           string([]rune{8214}),
	"\\textbrokenbar":        string([]rune{166}),
	"\\textreferencemark":    "\u203b",
	"\\textdiscount":         "\u2052",
	"\\textcelsius":          "\u2103",
	"\\textnumero":           string([]rune{8470}),
	"\\textrecipe":           string([]rune{8478}),
	"\\textestimated":        "\u212e",
	"\\textbigcircle":        string([]rune{9711}),
	"\\textmusicalnote":      string([]rune{9834}),
	"\\textohm":              "\u2126",
	"\\textmho":              "\u2127",
	"\\textleftarrow":        string([]rune{8592}),
	"\\textuparrow":          string([]rune{8593}),
	"\\textrightarrow":       string([]rune{8594}),
	"\\textdownarrow":        string([]rune{8595}),
	"\\textperthousand":      string([]rune{8240}),
	"\\textpertenthousand":   "\u2031",
	"\\textonehalf":          string([]rune{189}),
	"\\textthreequarters":    string([]rune{190}),
	"\\textonequarter":       string([]rune{188}),
	"\\textfractionsolidus":  string([]rune{8260}),
	"\\textdiv":              string([]rune{247}),
	"\\texttimes":            string([]rune{215}),
	"\\textminus":            string([]rune{8722}),
	"\\textasteriskcentered": string([]rune{8727}),
	"\\textpm":               string([]rune{177}),
	"\\textsurd":             string([]rune{8730}),
	"\\textlnot":             string([]rune{172}),
	"\\textonesuperior":      string([]rune{185}),
	"\\texttwosuperior":      string([]rune{178}),
	"\\textthreesuperior":    string([]rune{179}),
	"\\texteuro":             "€",
	"\\textcent":             "¢",
	"\\textsterling":         "£",
	"\\pounds":               "£",
	"\\textbaht":             "\u0e3f",
	"\\textcolonmonetary":    "\u20a1",
	"\\textcurrency":         "\u00a4",
	"\\textdong":             "\u20ab",
	"\\textflorin":           "\u0192",
	"\\textlira":             "\u20a4",
	"\\textnaira":            "\u20a6",
	"\\textpeso":             "\u20b1",
	"\\textwon":              "\u20a9",
	"\\textyen":              "\u00a5",
}
