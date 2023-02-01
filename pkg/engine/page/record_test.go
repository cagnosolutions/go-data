package page

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
		{0x11, R_NUM_NUM, true},
		{0x12, R_NUM_STR, true},
		{0x21, R_STR_NUM, true},
		{0x22, R_STR_STR, true},

		// pointer types
		{0x14, R_NUM_PTR, true},
		{0x24, R_STR_PTR, true},
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
