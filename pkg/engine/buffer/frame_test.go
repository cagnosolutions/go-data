package buffer

import (
	"testing"

	"github.com/cagnosolutions/go-data/pkg/engine/page"
)

func TestNewFrame(t *testing.T) {
	var f Frame
	if f.fid > 0 || f.pid > 0 || f.Page != nil {
		t.Errorf("new frame: frame should be nil")
	}
	f = newFrame(5, 3, page.PageSize)
	if f.fid == 0 || f.pid == 0 || f.Page == nil {
		t.Errorf("new frame: frame should not be nil")
	}
	_ = f
}

func TestFrame_DecrPinCount(t *testing.T) {
	f := newFrame(5, 3, page.PageSize)
	if f.fid == 0 || f.pid == 0 || f.Page == nil {
		t.Errorf("decr pin count: frame should not be nil")
	}
	f.pinCount += 2
	if f.pinCount != 3 {
		t.Errorf("decr pin count: bad pin count, should be 3, got %d", f.pinCount)
	}
	for i := 0; i < 25; i++ {
		f.decrPinCount()
	}
	if f.pinCount != 0 {
		t.Errorf("decr pin count: bad pin count, should be 0, got %d", f.pinCount)
	}
	_ = f
}

func TestFrame_Reset(t *testing.T) {
	f := newFrame(5, 3, page.PageSize)
	if f.fid == 0 || f.pid == 0 || f.Page == nil {
		t.Errorf("reset: frame should not be nil")
	}
	f.resetFrame()
	if f.fid > 0 || f.pid > 0 || f.Page != nil {
		t.Errorf("reset: frame should be nil")
	}
	_ = f
}
