package format

import (
	"fmt"
	"strings"
	"testing"
)

func TestDumper_Write(t *testing.T) {
	var sb strings.Builder
	dumper := NewDumper(&sb)
	dumper.Write([]byte("This is a test!!"))
	dumper.Write([]byte("A test of the dumper system!! :)"))
	dumper.Write([]byte("AAAAAAAAAAAAAAAA"))
	dumper.Write([]byte("AAAAAAAAAAAAAAAA"))
	dumper.Write([]byte("AAAAAAAAAAAAAAAA"))
	dumper.Write([]byte("Short."))
	dumper.Close()
	fmt.Println(sb.String())
}
