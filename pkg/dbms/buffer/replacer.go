package buffer

// Replacer represents a page replacement strategy.
type Replacer interface {

	// Victim removes the victim frame as defined by the replacement policy.
	Victim() *FrameID

	// Unpin unpins a frame, indicating that it can now be victimized.
	Unpin(id FrameID)

	// Pin pins a frame, indicating that it should not be victimized until it is unpinned.
	Pin(id FrameID)

	// Size returns the size of the replacer.
	Size() uint32

	// SetSize sets the replacement size.
	SetSize(size uint32)
}
