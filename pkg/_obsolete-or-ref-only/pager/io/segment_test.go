package io

/*
func _TestSegment(t *testing.T) {

	// create a temp directory
	dir, err := ioutil.TempDir(".", "segment-test-")
	if err != nil {
		t.Errorf("error creating temp directory")
	}

	// config
	const segSize = 1024
	const recSize = 64

	// initialize a new segment
	s, err := newSegment(dir, 1, segSize)
	if err != nil {
		t.Errorf("error creating new segment")
	}

}

func _TestSegment(t *testing.T) {

	// create a temp directory
	dir, _ := ioutil.TempDir(".", "segment-test-")
	defer os.RemoveAll(dir)

	// set up the config values
	const recSize = 64

	// want := &api.Record{Value: []byte("hello world")}
	want := make([]byte, recSize)
	copy(want, "hello world")

	// c := Config{}
	// c.Segment.MaxStoreBytes = 1024
	// c.Segment.MaxIndexBytes = entWidth * 3

	s, err := newSegment(dir, 64)
	util.AssertNoError(t, err)
	// util.AssertEqual(t, uint64(16), s.nextOffset, s.nextOffset)
	util.AssertTrue(t, !s.IsFull())

	var off uint64

	for i := uint64(0); i < 3; i++ {

		off = i * recSize

		_, err = s.WriteAt(want, int64(off))
		util.AssertNoError(t, err)
		util.AssertEqual(t, i*recSize, off)

		got := make([]byte, recSize)
		_, err = s.ReadAt(got, int64(off))

		util.AssertNoError(t, err)
		util.AssertEqual(t, want, got)
	}

	_, err = s.WriteAt(want, int64(off))
	util.AssertEqual(t, io.EOF, err)

	// maxed index
	util.AssertTrue(t, !s.IsFull())

	s, err = newSegment(dir, 16)
	util.AssertNoError(t, err)
	// maxed store
	util.AssertTrue(t, !s.IsFull())

	err = s.Remove()
	util.AssertNoError(t, err)

	s, err = newSegment(dir, 16)
	util.AssertNoError(t, err)
	util.AssertTrue(t, !s.IsFull())
}

*/
