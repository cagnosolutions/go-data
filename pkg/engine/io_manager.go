package engine

type IOManagerer interface {
	AllocatePage() PageID
	DeallocatePage(pid PageID) error
	ReadPage(pid PageID, p Page) error
	WritePage(pid PageID, p Page) error
	Close() error
}

type FileManager struct {
}
