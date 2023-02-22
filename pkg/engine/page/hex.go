package page

import (
	"bytes"
	"encoding/hex"
	"io"
	"log"
)

func HexDumpSkipDuplicates(p []byte) string {
	var buf bytes.Buffer
	d := hex.Dumper(&buf)
	defer func(d io.WriteCloser) {
		err := d.Close()
		if err != nil {
			log.Panicf("dumper close: %s", err)
		}
	}(d)
	// iterate the slice in sections of 16 bytes
	for len(p)%16 != 0 {
		p = append(p, 0x00)
	}
	for i := 0; i < len(p); i += 16 {
		d.Write(p[i : i+16])
		buf.WriteByte('\n')
	}
	// split each line out into a list of lines
	lines := bytes.Split(buf.Bytes(), []byte{'\n', '\n'})
	// range list of lines and make a list of duplicate
	// lines that we can skip in the next step
	x := make([]int, 0)
	const beg, end = 10, 58
	for i := range lines {
		if i < 1 {
			continue
		}
		if bytes.Equal(lines[i-1][beg:end], lines[i][beg:end]) {
			x = append(x, i)
		}
	}
	// reduce our skip-able lines to ranges
	skip := getranges(x)
	buf.Reset()
	// range our lines again, writing back to the buffer
	// any lines that should not be skipped
	var j int
	for i := range lines {
		if len(skip) > 0 {
			if len(skip[j]) == 1 && skip[j][0] == i {
				if j < len(skip) {
					buf.Write([]byte{'*', '\n'})
					j++
				}
				continue
			}
			if len(skip[j]) == 2 {
				if skip[j][0] == i {
					buf.Write([]byte{'*', '\n'})
				}
				if skip[j][0] <= i && skip[j][1] >= i {
					continue
				}
			}
		}
		buf.Write(lines[i])
		buf.WriteByte('\n')
	}
	return buf.String()
}

func getranges(s []int) [][]int {
	var n1, n2 int
	var ret [][]int
	for {
		n2 = n1 + 1
		for n2 < len(s) && s[n2] == s[n2-1]+1 {
			n2++
		}
		var at []int
		at = append(at, s[n1])
		if n2 == n1+2 {
			at = append(at, s[n2-1])
		}
		if n2 > n1+2 {
			at = append(at, s[n2-1])
		}
		ret = append(ret, at)
		if n2 == len(s) {
			break
		}
		if s[n2] == s[n2-1] {
			// repeating value
		}
		if s[n2] < s[n2-1] {
			// sequence not ordered
		}
		n1 = n2
	}
	return ret
}
