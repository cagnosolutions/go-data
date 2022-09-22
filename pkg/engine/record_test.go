package engine

import (
	"fmt"
	"testing"
)

func TestRecord_Flags(t *testing.T) {
	tests := []struct {
		want  uint8
		got   uint8
		valid bool
	}{
		// plain types
		{0x11, rNumNum, true},
		{0x12, rNumStr, true},
		{0x21, rStrNum, true},
		{0x22, rStrStr, true},

		// pointer types
		{0x14, rNumPtr, true},
		{0x24, rStrPtr, true},
	}
	for i, tt := range tests {
		if tt.got != tt.want {
			fmt.Printf("[%d] WARN: got=0x%.4x, want=0x%.4x\n", i, tt.got, tt.want)
		}
		_, err := newRecordHeader(tt.got, 5, 26)
		if err != nil && tt.valid == true {
			t.Errorf("[%d] got error with flags (0x%.4x): %s\n", i, tt.got, err)
		}
	}
}
