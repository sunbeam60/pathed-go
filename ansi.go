package main

// ANSI escape codes for terminal styling
const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiUnderline = "\x1b[4m"
	ansiNoUnder   = "\x1b[24m"
	ansiRed       = "\x1b[31m"
	ansiGreen     = "\x1b[32m"
	ansiYellow    = "\x1b[33m"
	ansiBlue      = "\x1b[34m"
	ansiBgWhite   = "\x1b[47m"
	ansiBgGrey    = "\x1b[100m"
	ansiBgRed     = "\x1b[101m"   // light red background
	ansiBgGreen   = "\x1b[102m"   // light green background
)
