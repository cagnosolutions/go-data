package pager

import (
	"fmt"
	"testing"
)

const (
	flg1 = 0x0001
	flg2 = 0x0002
	flg3 = 0x0004

	flgA = 0x0010
	flgB = 0x0020
	flgC = 0x0040
	flgD = 0x0080
)

func TestPage_addRecord(t *testing.T) {
	pg := newPage(1)
	id1, err := pg.addRecord([]byte("foo bar baz"))
	if err != nil {
		panic(err)
	}
	_ = id1
	fmt.Println(pg)
}

func fmtFlag(f uint16, flags ...uint16) string {
	ss := fmt.Sprintf("0x%.4x", f)
	if len(flags) > 0 {
		ss += " (flags set: "
		for i := range flags {
			ss += fmt.Sprintf("0x%.4x", flags[i])
			if i < len(flags)-1 {
				ss += ", "
			}
		}
		ss += ")"
	}
	return ss
}

type flgs []uint16

func checkingFlags(t *testing.T, round int, f uint16, flags ...uint16) {
	fmt.Printf("[%.3d] Checking flags:\t%s\n", round, fmtFlag(f, flags...))
}

func testFlags(t *testing.T, f uint16, yes []uint16, no []uint16) {
	if yes != nil && len(yes) > 0 {
		for _, flg := range yes {
			if !checkFlag(f, flg) {
				t.Errorf("Flag 0x%.4x IS NOT set, but it SHOULD be\n", flg)
			}
		}
	}
	if no != nil && len(no) > 0 {
		for _, flg := range no {
			if checkFlag(f, flg) {
				t.Errorf("Flag 0x%.4x IS set, but it SHOULD NOT be\n", flg)
			}
		}
	}
}

func TestFlags(t *testing.T) {

	// init round
	round := 1

	// round 01
	// init flags and check
	var flags uint16
	checkingFlags(t, round, flags)
	testFlags(t, flags, nil, flgs{flg1, flg2, flg3, flgA, flgB, flgC, flgD})
	round++

	// round 02
	// set flag 1 and check
	setFlag(&flags, flg1)
	checkingFlags(t, round, flags, flg1)
	testFlags(t, flags, flgs{flg1}, flgs{flg2, flg3, flgA, flgB, flgC, flgD})
	round++

	// round 03
	// unset flag 1 and set flag 2 and check
	unsetFlag(&flags, flg1)
	setFlag(&flags, flg2)
	checkingFlags(t, round, flags, flg2)
	testFlags(t, flags, flgs{flg2}, flgs{flg1, flg3, flgA, flgB, flgC, flgD})
	round++

	// round 04
	// unset flag 2 and set flag 3 and check
	unsetFlag(&flags, flg2)
	setFlag(&flags, flg3)
	checkingFlags(t, round, flags, flg3)
	testFlags(t, flags, flgs{flg3}, flgs{flg1, flg2, flgA, flgB, flgC, flgD})
	round++

	// round 05
	// unset flag 3 and set flag A and check
	unsetFlag(&flags, flg3)
	setFlag(&flags, flgA)
	checkingFlags(t, round, flags, flgA)
	testFlags(t, flags, flgs{flgA}, flgs{flg1, flg2, flg3, flgB, flgC, flgD})
	round++

	// round 06
	// unset flag A and set flag B and check
	unsetFlag(&flags, flgA)
	setFlag(&flags, flgB)
	checkingFlags(t, round, flags, flgB)
	testFlags(t, flags, flgs{flgB}, flgs{flg1, flg2, flg3, flgA, flgC, flgD})
	round++

	// round 07
	// unset flag B and set flag C and check
	unsetFlag(&flags, flgB)
	setFlag(&flags, flgC)
	checkingFlags(t, round, flags, flgC)
	testFlags(t, flags, flgs{flgC}, flgs{flg1, flg2, flg3, flgA, flgB, flgD})
	round++

	// round 08
	// unset flag C and set flag D and check
	unsetFlag(&flags, flgC)
	setFlag(&flags, flgD)
	checkingFlags(t, round, flags, flgD)
	testFlags(t, flags, flgs{flgD}, flgs{flg1, flg2, flg3, flgA, flgB, flgC})
	round++

	// round 09
	// unset flag D and check again
	unsetFlag(&flags, flgD)
	checkingFlags(t, round, flags)
	testFlags(t, flags, nil, flgs{flg1, flg2, flg3, flgA, flgB, flgC, flgD})
	round++

	// round 10
	// set flags 1 and 2 and check
	setFlag(&flags, flg1)
	setFlag(&flags, flg2)
	checkingFlags(t, round, flags, flg1, flg2)
	testFlags(t, flags, flgs{flg1, flg2}, flgs{flg3, flgA, flgB, flgC, flgD})
	round++

	// round 11
	// unset flags 1 and 2 and set flags 1 and 3 and check
	unsetFlag(&flags, flg1)
	unsetFlag(&flags, flg2)
	setFlag(&flags, flg1)
	setFlag(&flags, flg3)
	checkingFlags(t, round, flags, flg1, flg3)
	testFlags(t, flags, flgs{flg1, flg3}, flgs{flg2, flgA, flgB, flgC, flgD})
	round++

	// round 12
	// unset flags 1 and 3 and set flags A, B and C and check
	unsetFlag(&flags, flg1)
	unsetFlag(&flags, flg3)
	setFlag(&flags, flgA)
	setFlag(&flags, flgB)
	setFlag(&flags, flgC)
	checkingFlags(t, round, flags, flgA, flgB, flgC)
	testFlags(t, flags, flgs{flgA, flgB, flgC}, flgs{flg1, flg2, flg3, flgD})
	round++

	// round 13
	// unset flags A,B,C and set flags 1,2,3 and D
	unsetFlag(&flags, flgA)
	unsetFlag(&flags, flgB)
	unsetFlag(&flags, flgC)
	setFlag(&flags, flg1)
	setFlag(&flags, flg2)
	setFlag(&flags, flg3)
	setFlag(&flags, flgD)
	checkingFlags(t, round, flags, flg1, flg2, flg3, flgD)
	testFlags(t, flags, flgs{flg1, flg2, flg3, flgD}, flgs{flgA, flgB, flgC})
	round++

	if !checkFlags(flags, flg1, flg2, flg3, flgD) {
		t.Errorf("FAIL 1")
	}
	if checkFlags(flags, flg1, flg2, flg3, flgD, flgB) {
		t.Errorf("FAIL 2")
	}
	if !checkFlags(flags, flg2, flg3) {
		t.Errorf("FAIL 3")
	}

	// round 14
	// unset flags 1,2,3 and D and set flags A,B,C and D
	unsetFlag(&flags, flg1)
	unsetFlag(&flags, flg2)
	unsetFlag(&flags, flg3)
	unsetFlag(&flags, flgD)
	// setFlag(&flags, flgA)
	setFlag(&flags, flgB)
	setFlag(&flags, flgC)
	setFlag(&flags, flgD)
	checkingFlags(t, round, flags, flgB, flgC, flgD)
	testFlags(t, flags, flgs{flgB, flgC, flgD}, flgs{flg1, flg2, flg3})
	round++
}
