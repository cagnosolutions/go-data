package buffer

// Replacer is an interface describing the basic operations that make up a replacement
// policy. The replacer is used by the bufferPoolManager.
type Replacer interface {
	// Pin pins the frame matching the supplied frame ID, indicating that it should
	// not be victimized until it is unpinned.
	Pin(fid FrameID)
	// Victim removes and returns the next "victim frame", as defined by the policy.
	Victim() *FrameID
	// Unpin unpins the frame matching the supplied frame ID, indicating that it may
	// now be victimized.
	Unpin(fid FrameID)
	// Size should return the current number of elements in the replacer
	Size() uint16
}
