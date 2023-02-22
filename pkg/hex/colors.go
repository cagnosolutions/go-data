package hex

import (
	"fmt"
	"strconv"
	"strings"
)

const (

	// ESC is an escape character
	ESC = "\033"

	// RESET sets the color space back to "normal"
	RESET = ESC + "[m"
)

// Color codes
const (
	codeRed    = 1
	codeGreen  = 35
	codeYellow = 220
	codeBlue   = 33
	codePurple = 97
	codeOrange = 208
	codeGrey   = 246
)

const (
	red    = "\033[38;5;1m"
	green  = "\033[38;5;35m"
	yellow = "\033[38;5;220m"
	blue   = "\033[38;5;33m"
	pruple = "\033[38;5;97m"
	orange = "\033[38;5;208m"
	grey   = "\033[38;5;246m"
	reset  = "\033[m"
)

// Color strings
type Color string

const (

	// strRed is for non-printable ASCII characters
	strRed Color = ESC + "[38;5;1m"

	// strGreen is for ASCII whitespace characters
	strGreen Color = ESC + "[38;5;35m"

	// strYellow is for base 10 numerical digits
	strYellow Color = ESC + "[38;5;220m"

	// strBlue is currently not reserved for anything
	strBlue Color = ESC + "[38;5;33m"

	// strPurple is for punctuation characters
	strPurple Color = ESC + "[38;5;97m"

	// strOrange is for printable (alphabetic) ASCII characters
	strOrange Color = ESC + "[38;5;208m"

	// strGrey is for NULL byte
	strGrey Color = ESC + "[38;5;246m"
)

func PrintColor(color Color, s string) {
	fmt.Printf("%s%s", color, s)
}

func PrintfColor(color Color, format string, args ...any) {
	var sb strings.Builder
	sb.WriteString(string(color))
	sb.WriteString(format)
	fmt.Printf(sb.String(), args...)
}

func ASCIIColorChart() {
	var i, j, n int

	for i = 0; i < 11; i++ {
		for j = 0; j < 10; j++ {
			n = 10*i + j
			if n > 108 {
				break
			}
			fmt.Printf("\033[%dm %3d\033[m", n, n)
		}
		fmt.Printf("\n")
	}
}

func PrintFG(color int, s string) {
	var sb strings.Builder
	sb.WriteString(ESC)
	sb.WriteString("[38;5;")
	sb.WriteString(strconv.Itoa(color))
	sb.WriteString("m")
	sb.WriteString(s)
	sb.WriteString(RESET)
	fmt.Print(sb.String())

	// fmt.Printf("%s[38;5;%dm %s%s", ESC, n, s, RESET)
}

func PrintBG(n int, s string) {
	fmt.Printf("%s[48;5;%dm %s%s", ESC, n, s, RESET)
}

func ASCIIColorChart2() {

	var i int

	// Standard colors
	fmt.Printf("Standard Colors:\n")
	for i = 0; i < 8; i++ {
		PrintFG(i, fmt.Sprintf("%.3d", i))
	}
	fmt.Printf("\n\n")

	// High-intensity colors
	fmt.Printf("High-indensity Colors:\n")
	for i = 8; i < 16; i++ {
		PrintFG(i, fmt.Sprintf("%.3d", i))
	}
	fmt.Printf("\n\n")

	// More colors
	fmt.Printf("More Colors:\n")
	var j int
	for i = 16; i < 232; i++ {
		PrintFG(i, fmt.Sprintf("%.3d", i))
		if j == 5 {
			fmt.Printf("\n")
			j = 0
			continue
		}
		j++
	}
	fmt.Printf("\n\n")

	// Greyscale colors
	fmt.Printf("Greyscale Colors:\n")
	j = 0
	for i = 232; i < 256; i++ {
		PrintFG(i, fmt.Sprintf("%.3d", i))
		if j == 5 {
			fmt.Printf("\n")
			j = 0
			continue
		}
		j++
	}
}
