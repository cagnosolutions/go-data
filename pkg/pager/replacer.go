package pager

type clockReplacer struct {
}

func newClockReplacer(size int) *clockReplacer {
	return &clockReplacer{}
}

func (c *clockReplacer) pin(fid frameID) {
	// TODO implement me
	panic("implement me")
}

func (c *clockReplacer) unpin(fid frameID) {
	// TODO implement me
	panic("implement me")
}

func (c *clockReplacer) victim() frameID {
	// TODO implement me
	panic("implement me")
}

func (c *clockReplacer) size() int64 {
	// TODO implement me
	panic("implement me")
}
