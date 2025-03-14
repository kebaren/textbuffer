package common

// CharCode 定义了常用的字符代码
const (
	// Null 空字符
	Null = 0
	// Backspace 退格字符 '\b'
	Backspace = 8
	// Tab 制表符 '\t'
	Tab = 9
	// LineFeed 换行符 '\n'
	LineFeed = 10
	// CarriageReturn 回车符 '\r'
	CarriageReturn = 13
	// Space 空格
	Space = 32

	// ExclamationMark 感叹号 '!'
	ExclamationMark = 33
	// DoubleQuote 双引号 '"'
	DoubleQuote = 34
	// Hash 井号 '#'
	Hash = 35
	// DollarSign 美元符号 '$'
	DollarSign = 36
	// PercentSign 百分号 '%'
	PercentSign = 37
	// Ampersand 与符号 '&'
	Ampersand = 38
	// SingleQuote 单引号 '''
	SingleQuote = 39
	// OpenParen 左括号 '('
	OpenParen = 40
	// CloseParen 右括号 ')'
	CloseParen = 41
	// Asterisk 星号 '*'
	Asterisk = 42
	// Plus 加号 '+'
	Plus = 43
	// Comma 逗号 ','
	Comma = 44
	// Dash 破折号 '-'
	Dash = 45
	// Period 句号 '.'
	Period = 46
	// Slash 斜杠 '/'
	Slash = 47

	// 数字
	Digit0 = 48
	Digit1 = 49
	Digit2 = 50
	Digit3 = 51
	Digit4 = 52
	Digit5 = 53
	Digit6 = 54
	Digit7 = 55
	Digit8 = 56
	Digit9 = 57

	// Colon 冒号 ':'
	Colon = 58
	// Semicolon 分号 ';'
	Semicolon = 59
	// LessThan 小于号 '<'
	LessThan = 60
	// Equals 等号 '='
	Equals = 61
	// GreaterThan 大于号 '>'
	GreaterThan = 62
	// QuestionMark 问号 '?'
	QuestionMark = 63
	// AtSign 艾特符号 '@'
	AtSign = 64

	// 大写字母
	A = 65
	B = 66
	C = 67
	D = 68
	E = 69
	F = 70
	G = 71
	H = 72
	I = 73
	J = 74
	K = 75
	L = 76
	M = 77
	N = 78
	O = 79
	P = 80
	Q = 81
	R = 82
	S = 83
	T = 84
	U = 85
	V = 86
	W = 87
	X = 88
	Y = 89
	Z = 90

	// OpenSquareBracket 左方括号 '['
	OpenSquareBracket = 91
	// Backslash 反斜杠 '\'
	Backslash = 92
	// CloseSquareBracket 右方括号 ']'
	CloseSquareBracket = 93
	// Caret 脱字符 '^'
	Caret = 94
	// Underline 下划线 '_'
	Underline = 95
	// BackTick 反引号 '`'
	BackTick = 96

	// 小写字母
	a = 97
	b = 98
	c = 99
	d = 100
	e = 101
	f = 102
	g = 103
	h = 104
	i = 105
	j = 106
	k = 107
	l = 108
	m = 109
	n = 110
	o = 111
	p = 112
	q = 113
	r = 114
	s = 115
	t = 116
	u = 117
	v = 118
	w = 119
	x = 120
	y = 121
	z = 122

	// OpenCurlyBrace 左花括号 '{'
	OpenCurlyBrace = 123
	// Pipe 竖线 '|'
	Pipe = 124
	// CloseCurlyBrace 右花括号 '}'
	CloseCurlyBrace = 125
	// Tilde 波浪号 '~'
	Tilde = 126
)

// 组合字符和变音符
const (
	U_Combining_Grave_Accent               = 0x0300 // U+0300 Combining Grave Accent
	U_Combining_Acute_Accent               = 0x0301 // U+0301 Combining Acute Accent
	U_Combining_Circumflex_Accent          = 0x0302 // U+0302 Combining Circumflex Accent
	U_Combining_Tilde                      = 0x0303 // U+0303 Combining Tilde
	U_Combining_Macron                     = 0x0304 // U+0304 Combining Macron
	U_Combining_Overline                   = 0x0305 // U+0305 Combining Overline
	U_Combining_Breve                      = 0x0306 // U+0306 Combining Breve
	U_Combining_Dot_Above                  = 0x0307 // U+0307 Combining Dot Above
	U_Combining_Diaeresis                  = 0x0308 // U+0308 Combining Diaeresis
	U_Combining_Hook_Above                 = 0x0309 // U+0309 Combining Hook Above
	U_Combining_Ring_Above                 = 0x030A // U+030A Combining Ring Above
	U_Combining_Double_Acute_Accent        = 0x030B // U+030B Combining Double Acute Accent
	U_Combining_Caron                      = 0x030C // U+030C Combining Caron
	U_Combining_Vertical_Line_Above        = 0x030D // U+030D Combining Vertical Line Above
	U_Combining_Double_Vertical_Line_Above = 0x030E // U+030E Combining Double Vertical Line Above
	U_Combining_Double_Grave_Accent        = 0x030F // U+030F Combining Double Grave Accent
	U_Combining_Candrabindu                = 0x0310 // U+0310 Combining Candrabindu
	U_Combining_Inverted_Breve             = 0x0311 // U+0311 Combining Inverted Breve
	U_Combining_Turned_Comma_Above         = 0x0312 // U+0312 Combining Turned Comma Above
	U_Combining_Comma_Above                = 0x0313 // U+0313 Combining Comma Above
	U_Combining_Reversed_Comma_Above       = 0x0314 // U+0314 Combining Reversed Comma Above
	U_Combining_Comma_Above_Right          = 0x0315 // U+0315 Combining Comma Above Right
	U_Combining_Grave_Accent_Below         = 0x0316 // U+0316 Combining Grave Accent Below
	U_Combining_Acute_Accent_Below         = 0x0317 // U+0317 Combining Acute Accent Below
	U_Combining_Left_Tack_Below            = 0x0318 // U+0318 Combining Left Tack Below
	U_Combining_Right_Tack_Below           = 0x0319 // U+0319 Combining Right Tack Below
	U_Combining_Left_Angle_Above           = 0x031A // U+031A Combining Left Angle Above
	U_Combining_Horn                       = 0x031B // U+031B Combining Horn
	U_Combining_Left_Half_Ring_Below       = 0x031C // U+031C Combining Left Half Ring Below
	U_Combining_Up_Tack_Below              = 0x031D // U+031D Combining Up Tack Below
	U_Combining_Down_Tack_Below            = 0x031E // U+031E Combining Down Tack Below
	U_Combining_Plus_Sign_Below            = 0x031F // U+031F Combining Plus Sign Below
	U_Combining_Minus_Sign_Below           = 0x0320 // U+0320 Combining Minus Sign Below
)

// 更多组合字符
const (
	U_Combining_Palatalized_Hook_Below     = 0x0321 // U+0321 Combining Palatalized Hook Below
	U_Combining_Retroflex_Hook_Below       = 0x0322 // U+0322 Combining Retroflex Hook Below
	U_Combining_Dot_Below                  = 0x0323 // U+0323 Combining Dot Below
	U_Combining_Diaeresis_Below            = 0x0324 // U+0324 Combining Diaeresis Below
	U_Combining_Ring_Below                 = 0x0325 // U+0325 Combining Ring Below
	U_Combining_Comma_Below                = 0x0326 // U+0326 Combining Comma Below
	U_Combining_Cedilla                    = 0x0327 // U+0327 Combining Cedilla
	U_Combining_Ogonek                     = 0x0328 // U+0328 Combining Ogonek
	U_Combining_Vertical_Line_Below        = 0x0329 // U+0329 Combining Vertical Line Below
	U_Combining_Bridge_Below               = 0x032A // U+032A Combining Bridge Below
	U_Combining_Inverted_Double_Arch_Below = 0x032B // U+032B Combining Inverted Double Arch Below
	U_Combining_Caron_Below                = 0x032C // U+032C Combining Caron Below
	U_Combining_Circumflex_Accent_Below    = 0x032D // U+032D Combining Circumflex Accent Below
	U_Combining_Breve_Below                = 0x032E // U+032E Combining Breve Below
	U_Combining_Inverted_Breve_Below       = 0x032F // U+032F Combining Inverted Breve Below
	U_Combining_Tilde_Below                = 0x0330 // U+0330 Combining Tilde Below
	U_Combining_Macron_Below               = 0x0331 // U+0331 Combining Macron Below
	U_Combining_Low_Line                   = 0x0332 // U+0332 Combining Low Line
	U_Combining_Double_Low_Line            = 0x0333 // U+0333 Combining Double Low Line
	U_Combining_Tilde_Overlay              = 0x0334 // U+0334 Combining Tilde Overlay
	U_Combining_Short_Stroke_Overlay       = 0x0335 // U+0335 Combining Short Stroke Overlay
	U_Combining_Long_Stroke_Overlay        = 0x0336 // U+0336 Combining Long Stroke Overlay
	U_Combining_Short_Solidus_Overlay      = 0x0337 // U+0337 Combining Short Solidus Overlay
	U_Combining_Long_Solidus_Overlay       = 0x0338 // U+0338 Combining Long Solidus Overlay
	U_Combining_Right_Half_Ring_Below      = 0x0339 // U+0339 Combining Right Half Ring Below
	U_Combining_Inverted_Bridge_Below      = 0x033A // U+033A Combining Inverted Bridge Below
	U_Combining_Square_Below               = 0x033B // U+033B Combining Square Below
	U_Combining_Seagull_Below              = 0x033C // U+033C Combining Seagull Below
	U_Combining_X_Above                    = 0x033D // U+033D Combining X Above
	U_Combining_Vertical_Tilde             = 0x033E // U+033E Combining Vertical Tilde
	U_Combining_Double_Overline            = 0x033F // U+033F Combining Double Overline

	// 希腊字符和其他特殊字符
	LINE_SEPARATOR_2028 = 8232 // U+2028 LINE SEPARATOR
)

// 修饰符和特殊字符
const (
	// LineSeparator Unicode 行分隔符 (U+2028)
	LineSeparator = 8232
	// UTF8BOM UTF-8 BOM (U+FEFF)
	UTF8BOM = 65279

	// 其他常用 Unicode 字符
	U_CIRCUMFLEX          = 0x005E // U+005E CIRCUMFLEX
	U_GRAVE_ACCENT        = 0x0060 // U+0060 GRAVE ACCENT
	U_DIAERESIS           = 0x00A8 // U+00A8 DIAERESIS
	U_MACRON              = 0x00AF // U+00AF MACRON
	U_ACUTE_ACCENT        = 0x00B4 // U+00B4 ACUTE ACCENT
	U_CEDILLA             = 0x00B8 // U+00B8 CEDILLA
	U_SMALL_TILDE         = 0x02DC // U+02DC SMALL TILDE
	U_DOUBLE_ACUTE_ACCENT = 0x02DD // U+02DD DOUBLE ACUTE ACCENT
	U_OVERLINE            = 0x203E // U+203E OVERLINE
)

// 更多组合字符和希腊字符
const (
	U_Combining_Grave_Tone_Mark            = 0x0340 // U+0340 Combining Grave Tone Mark
	U_Combining_Acute_Tone_Mark            = 0x0341 // U+0341 Combining Acute Tone Mark
	U_Combining_Greek_Perispomeni          = 0x0342 // U+0342 Combining Greek Perispomeni
	U_Combining_Greek_Koronis              = 0x0343 // U+0343 Combining Greek Koronis
	U_Combining_Greek_Dialytika_Tonos      = 0x0344 // U+0344 Combining Greek Dialytika Tonos
	U_Combining_Greek_Ypogegrammeni        = 0x0345 // U+0345 Combining Greek Ypogegrammeni
	U_Combining_Bridge_Above               = 0x0346 // U+0346 Combining Bridge Above
	U_Combining_Equals_Sign_Below          = 0x0347 // U+0347 Combining Equals Sign Below
	U_Combining_Double_Vertical_Line_Below = 0x0348 // U+0348 Combining Double Vertical Line Below
	U_Combining_Left_Angle_Below           = 0x0349 // U+0349 Combining Left Angle Below
	U_Combining_Not_Tilde_Above            = 0x034A // U+034A Combining Not Tilde Above
	U_Combining_Homothetic_Above           = 0x034B // U+034B Combining Homothetic Above
	U_Combining_Almost_Equal_To_Above      = 0x034C // U+034C Combining Almost Equal To Above
	U_Combining_Left_Right_Arrow_Below     = 0x034D // U+034D Combining Left Right Arrow Below
	U_Combining_Upwards_Arrow_Below        = 0x034E // U+034E Combining Upwards Arrow Below
	U_Combining_Grapheme_Joiner            = 0x034F // U+034F Combining Grapheme Joiner

	// 修饰符字母和符号
	U_MODIFIER_LETTER_LEFT_ARROWHEAD  = 0x02C2 // U+02C2 MODIFIER LETTER LEFT ARROWHEAD
	U_MODIFIER_LETTER_RIGHT_ARROWHEAD = 0x02C3 // U+02C3 MODIFIER LETTER RIGHT ARROWHEAD
	U_MODIFIER_LETTER_UP_ARROWHEAD    = 0x02C4 // U+02C4 MODIFIER LETTER UP ARROWHEAD
	U_MODIFIER_LETTER_DOWN_ARROWHEAD  = 0x02C5 // U+02C5 MODIFIER LETTER DOWN ARROWHEAD
	U_BREVE                           = 0x02D8 // U+02D8 BREVE
	U_DOT_ABOVE                       = 0x02D9 // U+02D9 DOT ABOVE
	U_RING_ABOVE                      = 0x02DA // U+02DA RING ABOVE
	U_OGONEK                          = 0x02DB // U+02DB OGONEK

	// 希腊字符
	U_GREEK_TONOS                     = 0x0384 // U+0384 GREEK TONOS
	U_GREEK_DIALYTIKA_TONOS           = 0x0385 // U+0385 GREEK DIALYTIKA TONOS
	U_GREEK_KORONIS                   = 0x1FBD // U+1FBD GREEK KORONIS
	U_GREEK_PSILI                     = 0x1FBF // U+1FBF GREEK PSILI
	U_GREEK_PERISPOMENI               = 0x1FC0 // U+1FC0 GREEK PERISPOMENI
	U_GREEK_DIALYTIKA_AND_PERISPOMENI = 0x1FC1 // U+1FC1 GREEK DIALYTIKA AND PERISPOMENI
	U_GREEK_OXIA                      = 0x1FFD // U+1FFD GREEK OXIA
	U_GREEK_DASIA                     = 0x1FFE // U+1FFE GREEK DASIA
)
