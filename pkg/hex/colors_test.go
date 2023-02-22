package hex

import (
	"fmt"
	"testing"
)

func TestPrintColors(t *testing.T) {
	PrintFG(codeRed, "Red is for non-printable ASCII characters\n")
	fmt.Println("foo bar")
	PrintColor(strOrange, "Orange is for printable (alphabetic) characters\n")
	PrintColor(strYellow, "Yellow is for base 10 numerical digits\n")
	PrintColor(strGreen, "Green is for whitespace characters\n")
	PrintColor(strBlue, "Blue is a nice color\n")
	PrintColor(strPurple, "Purple is for punctuation characters\n")
	PrintColor(strGrey, "Grey is for NULL bytes\n")

}

func TestASCIIColorChart(t *testing.T) {
	ASCIIColorChart()
}

func TestASCIIColorChart2(t *testing.T) {
	ASCIIColorChart2()
}
