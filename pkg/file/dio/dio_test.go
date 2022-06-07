package dio

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func directIO(dir string) error {
	// Make a temporary file name
	fd, err := ioutil.TempFile("", dir)
	if err != nil {
		return fmt.Errorf("failed to make temp file: %q", err)
	}
	path := fd.Name()
	fd.Close()

	// starting block
	block1 := AlignedBlock(BlockSize)
	for i := 0; i < len(block1); i++ {
		block1[i] = 'A'
	}

	// Write the file
	out, err := OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to OpenFile for read: %q", err)
	}
	_, err = out.Write(block1)
	if err != nil {
		return fmt.Errorf("failed to write: %q", err)
	}
	err = out.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %q", err)
	}

	// Read the file
	block2 := AlignedBlock(BlockSize)
	in, err := OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to OpenFile for write: %q", err)
	}
	_, err = io.ReadFull(in, block2)
	if err != nil {
		return fmt.Errorf("failed to read: %q", err)
	}
	err = in.Close()
	if err != nil {
		return fmt.Errorf("failed to close reader: %q", err)
	}

	// Tidy
	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove temp file (%s): %q", path, err)
	}

	// Compare
	if !bytes.Equal(block1, block2) {
		return fmt.Errorf("read not the same as written")
	}

	return nil
}

func normalIO(dir string) error {
	// Make a temporary file name
	fd, err := ioutil.TempFile("", dir)
	if err != nil {
		return fmt.Errorf("failed to make temp file: %q", err)
	}
	path := fd.Name()
	fd.Close()

	// starting block
	block1 := AlignedBlock(BlockSize)
	for i := 0; i < len(block1); i++ {
		block1[i] = 'A'
	}

	// Write the file
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to OpenFile for read: %q", err)
	}
	_, err = out.Write(block1)
	if err != nil {
		return fmt.Errorf("failed to write: %q", err)
	}
	err = out.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %q", err)
	}

	// Read the file
	block2 := AlignedBlock(BlockSize)
	in, err := OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to OpenFile for write: %q", err)
	}
	_, err = io.ReadFull(in, block2)
	if err != nil {
		return fmt.Errorf("failed to read: %q", err)
	}
	err = in.Close()
	if err != nil {
		return fmt.Errorf("failed to close reader: %q", err)
	}

	// Tidy
	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove temp file (%s): %q", path, err)
	}

	// Compare
	if !bytes.Equal(block1, block2) {
		return fmt.Errorf("read not the same as written")
	}

	return nil
}

func TestDirectIO(t *testing.T) {
	// TIME IT!
	defer timeTrack(time.Now(), "DirectIO")
	err := directIO("direct_io_test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNormalIO(t *testing.T) {
	// TIME IT!
	defer timeTrack(time.Now(), "NormalIO")
	err := normalIO("normal_io_test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestZeroSizedBlock(t *testing.T) {
	// This should not panic!
	AlignedBlock(0)
}
