package buffer

import (
	"testing"

	"github.com/cagnosolutions/go-data/pkg/dbms/disk"
	"github.com/cagnosolutions/go-data/pkg/dbms/page"
	"github.com/cagnosolutions/go-data/pkg/util"
)

const poolSize = 10

func TestBufferPoolManager(t *testing.T) {

	bp := NewBufferPoolManager(poolSize, disk.NewDiskManagerTest())

	// Scenario: The buffer pool is empty. We should be able to create a new page.
	p0 := bp.NewPage()
	util.AssertEqual(t, page.PageID(0), p0.ID())

	pageData := [page.PageSize]byte{'h', 'e', 'l', 'l', 'o'}

	// Scenario: Once we have a page, we should be able to read and write content.
	p0.Copy(0, pageData[:])
	util.Equals(t, pageData, *p0.Data())

	// Scenario: We should be able to create new pages until we fill up the buffer pool.
	for i := uint32(1); i < poolSize; i++ {
		p := bp.NewPage()
		util.Equals(t, page.PageID(i), p.ID())
	}

	// Scenario: Once the buffer pool is full, we should not be able to create any new pages.
	for i := poolSize; i < poolSize*2; i++ {
		util.Equals(t, (*page.Page)(nil), bp.NewPage())
	}

	// Scenario: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		util.Ok(t, bp.UnpinPage(page.PageID(i), true))
		bp.FlushPage(page.PageID(i))
	}
	for i := 0; i < 4; i++ {
		bp.NewPage()
		// bp.UnpinPage(p.ID(), false)
	}

	// Scenario: We should be able to fetch the data we wrote a while ago.
	p0 = bp.FetchPage(page.PageID(0))
	util.Equals(t, pageData, *p0.Data())

	// Scenario: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	util.Ok(t, bp.UnpinPage(page.PageID(0), true))

	util.Equals(t, page.PageID(14), bp.NewPage().ID())
	util.Equals(t, (*page.Page)(nil), bp.NewPage())
	util.Equals(t, (*page.Page)(nil), bp.FetchPage(page.PageID(0)))
}
