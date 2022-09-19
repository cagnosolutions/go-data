package engine

import (
	"fmt"
	"testing"
)

const (
	rkNum = 0x0001 //	0000000000000001
	rkStr = 0x0002 //	0000000000000010
	rkPtr = 0x0004 // 	0000000000000100
	rkCus = 0x0008 // 	0000000000001000

	rvnum = 0x0010 // 	0000000000010000
	rvStr = 0x0020 // 	0000000000100000
	rvPtr = 0x0040 // 	0000000001000000
	rvCus = 0x0080 //  0000000010000000

	ptrRent = 0x0100 // 0000000100000000
	ptrChld = 0x0200 // 0000001000000000
	ptrPrev = 0x0400 // 0000010000000000
	ptrNext = 0x0800 // 0000100000000000

	_keyMask = 0x000f //  0000000000001111
	_valMask = 0x00f0 //  0000000011110000
	_ptrMask = 0x0f00 //  0000111100000000
	_cusMask = 0xf000 //  1111000000000000
)

func TestRecord_Flags(t *testing.T) {
	tests := []struct {
		want  uint32
		got   uint32
		valid bool
	}{
		// plain types
		{0x00000011, RK_NUM | RV_NUM, true},
		{0x00000021, RK_NUM | RV_STR, true},
		{0x00000012, RK_STR | RV_NUM, true},
		{0x00000022, RK_STR | RV_STR, true},

		// parent pointer types
		{0x00000141, PTR_RENT | RK_NUM | RV_PTR, true},
		{0x00000142, PTR_RENT | RK_STR | RV_PTR, true},

		// invalid mixins
		{0x00000111, PTR_RENT | RK_NUM | RV_NUM, false},
		{0x00000121, PTR_RENT | RK_NUM | RV_STR, false},

		// child pointer types
		{0x00000241, PTR_CHLD | RK_NUM | RV_PTR, true},
		{0x00000242, PTR_CHLD | RK_STR | RV_PTR, true},

		// invalid mixins
		{0x00000041, RK_NUM | RV_PTR, false},
		{0x00000042, RK_STR | RV_PTR, false},

		// prev pointer types
		{0x00000441, PTR_PREV | RK_NUM | RV_PTR, true},
		{0x00000442, PTR_PREV | RK_STR | RV_PTR, true},
		{0x00000441, PTR_PREV | RK_NUM | RV_PTR, true},

		// invalid mixins
		{0x00000141, RK_NUM | RV_PTR, false},
		{0x00000111, PTR_RENT | RK_NUM | RV_NUM, false},

		// next pointer types
		{0x00000841, PTR_NEXT | RK_NUM | RV_PTR, true},
		{0x00000842, PTR_NEXT | RK_STR | RV_PTR, true},

		// invalid mixins
		{0x00000121, PTR_RENT | RK_NUM | RV_STR, false},
	}
	for i, tt := range tests {
		if tt.got != tt.want {
			fmt.Printf("[%d] WARN: got=0x%.4x, want=0x%.4x\n", i, tt.got, tt.want)
		}
		err := checkRecordFlags(tt.got)
		if err != nil && tt.valid == true {
			t.Errorf("[%d] got error with flags (0x%.4x): %s\n", i, tt.got, err)
		}
	}
}
