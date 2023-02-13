package engine

import (
	"fmt"
	"testing"
)

var conf = struct {
	path       string
	namespaces []string
}{
	path:       "testing/db",
	namespaces: []string{"users", "orders"},
}

func write(format string, args ...any) []byte {
	return []byte(fmt.Sprintf(format, args...))
}

func TestStorageEngine_Namespaces(t *testing.T) {

	// Open the storage engine
	se, err := OpenStorageEngine(conf.path)
	if err != nil {
		t.Error(err)
	}

	// Create a namespace
	err = se.CreateNamespace(conf.namespaces[0])
	if err != nil {
		t.Error(err)
	}

	// Write some data to the namespace
	var id uint32
	var ids []uint32
	for i := 0; i < 10; i++ {
		id, err = se.Insert(conf.namespaces[0], write("[record-%d] some data for %q", i, conf.namespaces[0]))
		if err != nil {
			t.Error(err)
		}
		ids = append(ids, id)
	}
	err = se.Save()
	if err != nil {
		t.Error(err)
	}

	// Print information about the engine
	fmt.Println(se.Info())
	fmt.Println(se.NamespaceInfo(conf.namespaces[0]))

	// Close the storage engine
	err = se.Close()
	if err != nil {
		t.Error(err)
	}
}
