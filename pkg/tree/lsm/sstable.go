package lsm

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// SSTable represents a Sorted String Table using the WiscKey format
type SSTable struct {
	index *os.File
	data  *os.File
}

type Entry struct {
	Key string
	Val string
}

type IndexEntry struct {
	Key       string
	OffsetPtr uint32
}

type Entries []Entry

func (e Entries) Len() int           { return len(e) }
func (e Entries) Less(i, j int) bool { return e[i].Key < e[j].Key }
func (e Entries) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func OpenSSTable(baseDir string) (*SSTable, error) {

	// init variables
	var indexFile, dataFile *os.File
	// var existed bool

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		// baseDir does not exist, so we need to create
		baseDir, err = initBasePath(baseDir)
		if err != nil {
			return nil, err
		}
		// create new files
		indexFile, dataFile, err = createIndexAndDataFiles(baseDir)
		if err != nil {
			return nil, err
		}
		//	existed = false
	} else {
		// baseDir does exist, so we need to open
		indexFile, dataFile, err = openIndexAndDataFiles(filepath.ToSlash(baseDir))
		if err != nil {
			return nil, err
		}
		//	existed = true
	}

	// // Sort the entries by key to maintain order in the SSTable
	// sort.Stable(Entries(entries))
	//
	// if !existed {
	//
	// 	var index []uint32
	//
	// 	// Build the data file
	// 	for _, entry := range entries {
	//
	// 		// Get the current offset
	// 		offset, err := dataFile.Seek(0, io.SeekCurrent)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		// Store the offset in the index for the current key
	// 		index = append(index, uint32(offset))
	//
	// 		// Write the key length, key, value length and value to the data file
	// 		// keyLen, valLen := uint32(len(entry.Key)), uint32(len(entry.Val))
	//
	// 		// Write key length
	// 		err = binary.Write(dataFile, binary.BigEndian, uint32(len(entry.Key)))
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		// Write key
	// 		_, err = dataFile.WriteString(entry.Key)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		// Write value length
	// 		err = binary.Write(dataFile, binary.BigEndian, uint32(len(entry.Val)))
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		// Write value
	// 		_, err = dataFile.WriteString(entry.Val)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	}
	//
	// 	// Write the index to the index file
	// 	for _, offset := range index {
	// 		// Write key offsets
	// 		err := binary.Write(indexFile, binary.BigEndian, offset)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// }

	return &SSTable{
		index: indexFile,
		data:  dataFile,
	}, nil

}

func (s *SSTable) WriteBatch(entries []Entry) error {
	// Sort the entries by key to maintain order in the SSTable
	sort.Stable(Entries(entries))

	var index []uint32

	// Build the data file
	for _, entry := range entries {

		// Get the current offset
		offset, err := s.data.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		// Store the offset in the index for the current key
		index = append(index, uint32(offset))

		// Write the key length, key, value length and value to the data file
		// keyLen, valLen := uint32(len(entry.Key)), uint32(len(entry.Val))

		// Write key length
		err = binary.Write(s.data, binary.BigEndian, uint32(len(entry.Key)))
		if err != nil {
			return err
		}

		// Write key
		_, err = s.data.WriteString(entry.Key)
		if err != nil {
			return err
		}

		// Write value length
		err = binary.Write(s.data, binary.BigEndian, uint32(len(entry.Val)))
		if err != nil {
			return err
		}

		// Write value
		_, err = s.data.WriteString(entry.Val)
		if err != nil {
			return err
		}
	}

	// Write the index to the index file
	for _, offset := range index {
		// Write key offsets
		err := binary.Write(s.index, binary.BigEndian, offset)
		if err != nil {
			return err
		}
	}

	return nil
}

func createIndexAndDataFiles(basePath string) (*os.File, *os.File, error) {
	// Create files
	indexFile, err := os.Create(filepath.Join(basePath, "index.sst"))
	if err != nil {
		return nil, nil, err
	}
	dataFile, err := os.Create(filepath.Join(basePath, "data.sst"))
	if err != nil {
		return nil, nil, err
	}

	return indexFile, dataFile, nil
}

func openIndexAndDataFiles(basePath string) (*os.File, *os.File, error) {
	// Create files
	indexFile, err := os.OpenFile(filepath.Join(basePath, "index.sst"), os.O_RDWR, 0666)
	if err != nil {
		return nil, nil, err
	}
	dataFile, err := os.OpenFile(filepath.Join(basePath, "data.sst"), os.O_RDWR, 0666)
	if err != nil {
		return nil, nil, err
	}

	return indexFile, dataFile, nil
}

func (s *SSTable) GetBinary(key string) (string, bool, error) {
	// Read the index file to find the offset for the given key.
	offset, err := s.findIndexOffsetBinary(key)
	if err != nil {
		return "", false, err
	}
	if offset == -1 {
		return "", false, nil // Key not found.
	}

	// Read the entry from the data file using the offset.
	entry, err := s.readDataEntry(offset)
	if err != nil {
		return "", false, err
	}

	return entry.Val, true, nil
}

func (s *SSTable) Get(key string) (string, bool, error) {
	// Read the index file to find the offset for the given key.
	offset, err := s.findIndexOffset(key)
	if err != nil {
		return "", false, err
	}
	if offset == -1 {
		return "", false, nil // Key not found.
	}

	// Read the entry from the data file using the offset.
	entry, err := s.readDataEntry(offset)
	if err != nil {
		return "", false, err
	}

	return entry.Val, true, nil
}

func (s *SSTable) findIndexOffset(key string) (int64, error) {
	// Read the index file to find the offset for the given key.
	var offset uint32

	for {
		if err := binary.Read(s.index, binary.BigEndian, &offset); err != nil {
			if err == io.EOF {
				// End of index file, key not found.
				return -1, nil
			}
			return -1, err
		}

		// Read the key at the current offset.
		_, err := s.data.Seek(int64(offset), io.SeekStart)
		if err != nil {
			return -1, err
		}
		var entryKeyLen uint32
		if err := binary.Read(s.data, binary.BigEndian, &entryKeyLen); err != nil {
			return -1, err
		}

		entryKeyBytes := make([]byte, entryKeyLen)
		if _, err := s.data.Read(entryKeyBytes); err != nil {
			return -1, err
		}

		entryKey := string(entryKeyBytes)

		if entryKey == key {
			// Key found.
			return int64(offset), nil
		}
	}
}

func (s *SSTable) readDataEntry(offset int64) (*Entry, error) {
	// Read the data file at the given offset to get the entry.
	var entry Entry

	_, err := s.data.Seek(offset, io.SeekStart)
	if err != nil {
		return &entry, err
	}

	var keyLen, valueLen uint32
	if err := binary.Read(s.data, binary.BigEndian, &keyLen); err != nil {
		return &entry, err
	}

	keyBytes := make([]byte, keyLen)
	if _, err := s.data.Read(keyBytes); err != nil {
		return &entry, err
	}

	entry.Key = string(keyBytes)

	if err := binary.Read(s.data, binary.BigEndian, &valueLen); err != nil {
		return &entry, err
	}

	valueBytes := make([]byte, valueLen)
	if _, err := s.data.Read(valueBytes); err != nil {
		return &entry, err
	}

	entry.Val = string(valueBytes)

	return &entry, nil
}

func (s *SSTable) findIndexOffsetBinary(key string) (int64, error) {
	// Read the index file to find the offset for the given key using binary search.
	var offset uint32

	// Read the total number of entries from the index file.
	var numEntries uint32
	if err := binary.Read(s.index, binary.BigEndian, &numEntries); err != nil {
		return -1, err
	}

	// Perform binary search on the index.
	left, right := uint32(0), numEntries-1
	for left <= right {
		mid := (left + right) / 2

		// Read the offset at the middle index.
		if _, err := s.index.Seek(int64(mid*4+8), io.SeekStart); err != nil {
			return -1, err
		}
		if err := binary.Read(s.index, binary.BigEndian, &offset); err != nil {
			return -1, err
		}

		// Read the key at the current offset.
		if _, err := s.data.Seek(int64(offset), io.SeekStart); err != nil {
			return -1, err
		}
		var entryKeyLen uint32
		if err := binary.Read(s.data, binary.BigEndian, &entryKeyLen); err != nil {
			return -1, err
		}

		entryKeyBytes := make([]byte, entryKeyLen)
		if _, err := s.data.Read(entryKeyBytes); err != nil {
			return -1, err
		}

		entryKey := string(entryKeyBytes)

		// Compare the entry key with the given key.
		if entryKey == key {
			// Key found.
			return int64(offset), nil
		} else if entryKey < key {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	// Key not found.
	return -1, nil
}

func (s *SSTable) Close() error {
	err := s.index.Close()
	if err != nil {
		return err
	}
	err = s.data.Close()
	if err != nil {
		return err
	}
	return nil
}
