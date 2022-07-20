package pager

import "fmt"

func blockAlign(count uint8, block uint32) uint32 {
	size := uint32(count) * block
	for block < size {
		block *= 2
	}
	return block
}

const (
	szSector          = 512
	minPageSectors    = 4
	maxPageSectors    = 128
	minExtentPages    = 4
	maxExtentPages    = 64
	minSegmentExtents = 4
	maxSegmentExtents = 32
)

type pageInfo struct {
	sectors  uint8  // number of sectors in the page
	pageSize uint32 // total size of the page (in bytes)
}

func newPageInfo(sectors uint8) *pageInfo {
	if sectors < minPageSectors {
		sectors = minPageSectors
	}
	if sectors > maxPageSectors {
		sectors = maxPageSectors
	}
	return &pageInfo{
		sectors:  sectors,
		pageSize: blockAlign(sectors, szSector),
	}
}

type extentInfo struct {
	pages      uint8  // number of pages in the extent
	extentSize uint32 // total size of the extent (in bytes)
}

func newExtentInfo(pages uint8, pageSize uint32) *extentInfo {
	if pages < minExtentPages {
		pages = minExtentPages
	}
	if pages > maxExtentPages {
		pages = maxExtentPages
	}
	return &extentInfo{
		pages:      pages,
		extentSize: blockAlign(pages, pageSize),
	}
}

type segmentInfo struct {
	extents     uint8  // number of extents in the segment
	segmentSize uint32 // total size of the segment (in bytes)
}

func newSegmentInfo(extents uint8, extentSize uint32) *segmentInfo {
	if extents < minSegmentExtents {
		extents = minSegmentExtents
	}
	if extents > maxSegmentExtents {
		extents = maxSegmentExtents
	}
	return &segmentInfo{
		extents:     extents,
		segmentSize: blockAlign(extents, extentSize),
	}
}

type allocator struct {
	*pageInfo
	*extentInfo
	*segmentInfo
}

func newAllocator(pageSize uint32) *allocator {
	sectors := uint8(pageSize / szSector)
	if sectors > ^uint8(0) {
		sectors = ^uint8(0)
	}
	pi := newPageInfo(sectors)
	ei := newExtentInfo(pi.sectors>>4, pi.pageSize)
	si := newSegmentInfo(ei.pages>>2, ei.extentSize)
	return &allocator{
		pageInfo:    pi,
		extentInfo:  ei,
		segmentInfo: si,
	}
}

func newAllocatorSize(pageSize, extentSize, segmentSize uint32) *allocator {
	// calculate number of sectors for new page
	sectors := uint8(pageSize / szSector)
	if sectors > ^uint8(0) {
		sectors = ^uint8(0)
	}
	pi := newPageInfo(sectors)
	// calculate number of pages for new extent
	pages := uint8(extentSize / pi.pageSize)
	if pages > ^uint8(0) {
		pages = ^uint8(0)
	}
	ei := newExtentInfo(pages, pi.pageSize)
	// calculate number of extents for new segment
	extents := uint8(segmentSize / ei.extentSize)
	if extents > ^uint8(0) {
		extents = ^uint8(0)
	}
	si := newSegmentInfo(extents, ei.extentSize)
	return &allocator{
		pageInfo:    pi,
		extentInfo:  ei,
		segmentInfo: si,
	}
}

func (a *allocator) String() string {
	ss := fmt.Sprintf("allocator:\n")
	ss += fmt.Sprintf("\tpageInfo (%s):\n", formatSize(uint64(a.pageSize), true, false))
	ss += fmt.Sprintf("\t\tsectors=%d\n", a.sectors)
	ss += fmt.Sprintf("\t\tpageSize=%d\n", a.pageSize)
	ss += fmt.Sprintf("\textentInfo (%s):\n", formatSize(uint64(a.extentSize), true, false))
	ss += fmt.Sprintf("\t\tpages=%d\n", a.pages)
	ss += fmt.Sprintf("\t\textentSize=%d\n", a.extentSize)
	ss += fmt.Sprintf("\tsegmentInfo (%s):\n", formatSize(uint64(a.segmentSize), true, false))
	ss += fmt.Sprintf("\t\textents=%d\n", a.extents)
	ss += fmt.Sprintf("\t\tsegmentSize=%d\n", a.segmentSize)
	return ss
}

func formatSize(size uint64, useBin, useFP bool) string {
	const (
		b   = 1
		kib = 1 << 10 // kibibyte (1024 bytes)
		mib = 1 << 20 // mebibyte (1024 kibibyte)
		gib = 1 << 30 // gibibyte (1024 mebibyte)
		tib = 1 << 40 // tebibyte (1024 gibibyte)
		pib = 1 << 50 // pebibyte (1024 tebibyte)
		eib = 1 << 60 // exbibyte (1024 pebibyte)

		kb = kib - (24 * b)  // kilobyte (1000 bytes)
		mb = mib - (24 * kb) // megabyte (1000 kilobytes)
		gb = gib - (24 * mb) // gigabyte (1000 megabytes)
		tb = tib - (24 * gb) // terabyte (1000 gigabytes)
		pb = pib - (24 * tb) // petabyte (1000 terabytes)
		eb = eib - (24 * pb) // exabyte  (1000 petabytes)
	)
	suffix := map[uint64]string{
		b:   "bytes",
		kb:  "KB",
		mb:  "MB",
		gb:  "GB",
		tb:  "TB",
		pb:  "PB",
		eb:  "EB",
		kib: "KiB",
		mib: "MiB",
		gib: "GiB",
		tib: "TiB",
		pib: "PiB",
		eib: "EiB",
	}
	tooLargeStr := "Size is too large (in the zettabyte range) and can not be computed."
	var res uint64
	if useBin {
		switch {
		case size < kib:
			res = b
		case size > kib && size < mib:
			res = kib
		case size > mib && size < gib:
			res = mib
		case size > gib && size < tib:
			res = gib
		case size > tib && size < pib:
			res = tib
		case size > pib && size < eib:
			res = pib
		default:
			return tooLargeStr
		}
	} else {
		switch {
		case size < kb:
			res = b
		case size > kb && size < mb:
			res = kb
		case size > mb && size < gb:
			res = mb
		case size > gb && size < tb:
			res = gb
		case size > tb && size < pb:
			res = tb
		case size > pb && size < eb:
			res = pb
		default:
			return tooLargeStr
		}
	}
	if useFP && res != b {
		return fmt.Sprintf("%2.2f %s", float64(size)/float64(res), suffix[res])
	}
	return fmt.Sprintf("%d %s", size/res, suffix[res])
}
