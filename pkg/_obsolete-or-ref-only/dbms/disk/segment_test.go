package disk

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSegmentHeader_logicalPageOffset(t *testing.T) {

	// create new segment header
	s := makeSegmentHeader(1, pageSize)

	// get the logical offset of the segment header
	off, err := s.logicalPageOffset(5000)
	if err != nil {
		t.Error(err)
	}

	// print out logical offset
	fmt.Println(off, off == uint32(s.pageSize*5000) && off%uint32(s.pageSize) == 0)
}

func TestSegmentHeader_WriteAndRead(t *testing.T) {

	// create new segment header (segment 1)
	s1 := makeSegmentHeader(1, pageSize)

	// create a byte slice to write data into
	p := make([]byte, segmentHeaderSize)

	// write segment header into buffer
	n, err := s1.Write(p)
	if err != nil {
		t.Error(err)
	}
	if n != int(segmentHeaderSize) {
		t.Error("wrote incorrect segment header length")
	}

	fmt.Printf("p=%v\n", p)

	// create a new segment header (segment 2)
	s2 := makeSegmentHeader(4, pageSize)

	// read info from buffer into segment
	n, err = s2.Read(p)
	if err != nil {
		t.Error(err)
	}
	if n != int(segmentHeaderSize) {
		t.Error("read incorrect segment header length")
	}

	// compare segment 1 and segment 2
	if *s1 != *s2 {
		t.Error("segment 1 and segment 2 are not the same")
	}

	// print out segment 1 and segment 2
	fmt.Println(s1, s2)
}

func TestSegmentHeader_WriteToAndReadFrom(t *testing.T) {

	// create new segment header (segment 1)
	s1 := makeSegmentHeader(0, pageSize)

	// create a "file" to write data to
	tmpFile := new(bytes.Buffer)

	// write segment header to "file"
	n, err := s1.WriteTo(tmpFile)
	if err != nil {
		t.Error(err)
	}
	if int(n) != int(segmentHeaderSize) {
		t.Error("wrote incorrect segment header length")
	}

	fmt.Printf("tmpFile=%v\n", tmpFile.Bytes())

	// create a new segment header (segment 2)
	s2 := makeSegmentHeader(4, pageSize)

	// read from the "file" into our segment
	n, err = s2.ReadFrom(tmpFile)
	if err != nil {
		t.Error(err)
	}
	if int(n) != int(segmentHeaderSize) {
		t.Error("read incorrect segment header length")
	}

	// compare segment 1 and segment 2
	if *s1 != *s2 {
		t.Error("segment 1 and segment 2 are not the same")
	}

	// print out segment 1 and segment 2
	fmt.Println(s1, s2)
}
