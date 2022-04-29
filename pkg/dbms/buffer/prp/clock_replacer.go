package prp

// FrameID represents a page frame.
type FrameID uint32

// ClockReplacer represents the clock replacer algorithm
type ClockReplacer struct {
	cList     *circularList
	clockHand **node
}

// NewClockReplacer instantiates a new clock replacer
func NewClockReplacer(poolSize uint32) *ClockReplacer {
	cList := newCircularList(poolSize)
	return &ClockReplacer{cList, &cList.head}
}

// Victim removes the victim frame as defined by the replacement policy
func (c *ClockReplacer) Victim() *FrameID {
	if c.cList.size == 0 {
		return nil
	}

	var victimFrameID *FrameID
	currentNode := *(*c.clockHand)
	for {
		if currentNode.value.(bool) {
			currentNode.value = false
			c.clockHand = &currentNode.next
		} else {
			frameID := currentNode.key.(FrameID)
			victimFrameID = &frameID

			c.clockHand = &currentNode.next

			c.cList.remove(currentNode.key)
			return victimFrameID
		}
	}
}

// Unpin unpins a frame, indicating that it can now be victimized
func (c *ClockReplacer) Unpin(id FrameID) {
	if !c.cList.hasKey(id) {
		c.cList.insert(id, true)
		if c.cList.size == 1 {
			c.clockHand = &c.cList.head
		}
	}
}

// Pin pins a frame, indicating that it should not be victimized until it is unpinned
func (c *ClockReplacer) Pin(id FrameID) {
	n := c.cList.find(id)
	if n == nil {
		return
	}

	if (*c.clockHand) == n {
		c.clockHand = &(*c.clockHand).next
	}
	c.cList.remove(id)

}

// Size returns the size of the clock
func (c *ClockReplacer) Size() uint32 {
	return c.cList.size
}

func (c *ClockReplacer) SetSize(size uint32) {
	c.cList.size = size
}
