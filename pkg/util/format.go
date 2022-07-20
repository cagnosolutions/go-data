package util

import (
	"fmt"
)

func FormatSize(size uint64, useBin, useFP bool) string {
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
