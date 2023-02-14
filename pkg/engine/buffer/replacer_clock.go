package buffer

import (
	"errors"
	"log"
)

var (
	ErrListIsFull    = errors.New("circular-list: full; circular list capacity met")
	ErrNoVictimFound = errors.New("clock-replacer: may be empty; victim could not be found")
)

// ClockReplacer represents a clock based replacement cache
type ClockReplacer struct {
	list *circularList[FrameID, bool]
	ptr  **dllNode[FrameID, bool]
}

// newClockReplacer instantiates and returns a new ClockReplacer
func NewClockReplacer(size uint16) *ClockReplacer {
	list := newCircularList[FrameID, bool](size)
	return &ClockReplacer{
		list: list,
		ptr:  &list.head,
	}
}

// Pin takes a frame ID and "pins" it, indicating that the caller is now
// using it. Because the caller is now using it, the replacer can now remove
// it to no longer make it available for victimization.
func (c *ClockReplacer) Pin(fid FrameID) {
	n := c.list.find(fid)
	if n == nil {
		return
	}
	if (*c.ptr) == n {
		c.ptr = &(*c.ptr).next
	}
	c.list.remove(fid)
}

// Unpin takes a frame ID and "unpins" it, indicating that the caller is no
// longer using it. Because the caller is no longer using it, the replacer
// can now add it to make it available for victimization.
func (c *ClockReplacer) Unpin(fid FrameID) {
	if !c.list.hasKey(fid) {
		if err := c.list.insert(fid, true); err != nil {
			log.Panicf("replacer.unpin: failed on insert: %q", err)
		}
		if c.list.size == 1 {
			c.ptr = &c.list.head
		}
	}
}

// Victim searches for a frame ID in the replacer that it can victimize and
// return to the caller. It locates and removes a victim frame ID (as defined
// by the replacement policy) and returns it. If there are no frame IDs to
// victimize, it will simply return nil. In the case of a nil return, the caller
// will have to figure out how to handle the situation.
func (c *ClockReplacer) Victim() *FrameID {
	if c.list.size == 0 {
		return nil
	}
	var victim *FrameID
	cn := *c.ptr
	for {
		if cn.val {
			cn.val = false
			c.ptr = &cn.next
		} else {
			fid := cn.key
			victim = &fid
			c.ptr = &cn.next
			c.list.remove(cn.key)
			return victim
		}
	}
}

// size returns the number of elements currently in the replacer.
func (c *ClockReplacer) Size() uint16 {
	return c.list.size
}

// String is the stringer method for this type.
func (c *ClockReplacer) String() string {
	return c.list.String()
}
