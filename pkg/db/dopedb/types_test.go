package dopedb

import (
	"bytes"
	"strings"
	"testing"
)

func TestStrings(t *testing.T) {

	var strTests = []struct {
		str string
		len int
		exp byte
		bin []byte
	}{
		{
			"Testing",
			1 + 7,
			FixStr | 7,
			nil,
		},
		{
			"Testing, 1... 2... 3...",
			1 + 23,
			FixStr | 23,
			nil,
		},
		{
			"Testing, 1... 2... 3... This is a test of the emergency broadcast system.",
			1 + 1 + 73,
			Str8,
			nil,
		},
		{
			"Testing, 1... 2... 3... This is a test of the emergency broadcast system. This is only a test. " +
				"Please do not be concerned. Everything is just fine... Uhhh, " +
				"wait... I am getting some new information now.",
			1 + 1 + 202,
			Str8,
			nil,
		},
		{
			"Testing, 1... 2... 3... This is a test of the emergency broadcast system. This is only a test. " +
				"Please do not be concerned. Everything is just fine... Uhhh, " +
				"wait... I am getting some new information now. Okay--I don't want anyone to worry... " +
				"But it seems we have been invaded by an alien race.",
			1 + 2 + 292,
			Str16,
			nil,
		},
		{
			"Testing, 1... 2... 3... This is a test of the emergency broadcast system. This is only a test. " +
				"Please do not be concerned. Everything is just fine... Uhhh, " +
				"wait... I am getting some new information now. Okay--I don't want anyone to worry... " +
				"But it seems we have been invaded by an alien race. " +
				"They have come to every major city and have established some kind of military approach upon meeting" +
				" some people in government. I will be sure to report back as soon as I... AHhh...... shhhwoooshhh..." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"....................................................................................................",
			1 + 2 + 3692,
			Str16,
			nil,
		},
		{
			strings.Repeat(" [We have come to take over]", 1<<20),
			1 + 4 + 29360128,
			Str32,
			nil,
		},
		{
			strings.Repeat(" [We have come to take over]", 1<<21),
			1 + 4 + 58720256,
			Str32,
			nil,
		},
	}

	// testing string encoding
	for i, tt := range strTests {
		out, err := encStr(tt.str)
		if err != nil {
			t.Errorf("%s", err)
			continue
		}
		if out[0] != tt.exp {
			t.Errorf("error: type byte encoded incorrectly, got=%x, wanted=%x\n", out[0], tt.exp)
		}
		if len(out) != tt.len {
			t.Errorf("error: length of encoded output is incorrect, got=%d, wanted=%d\n", len(out), tt.len)
		}
		strTests[i].bin = out
	}

	// testing string decoding
	for _, tt := range strTests {
		out, err := decStr(tt.bin)
		if err != nil {
			t.Errorf("%s", err)
			continue
		}
		if out != tt.str {
			t.Errorf("error: decoded string did not match origional string\n")
		}
	}
}

func TestBinaryBytes(t *testing.T) {

	var binTests = []struct {
		str []byte
		len int
		exp byte
		bin []byte
	}{
		{
			[]byte("Testing"),
			1 + 1 + 7,
			Bin8,
			nil,
		},
		{
			[]byte("Testing, 1... 2... 3..."),
			1 + 1 + 23,
			Bin8,
			nil,
		},
		{
			[]byte("Testing, 1... 2... 3... This is a test of the emergency broadcast system."),
			1 + 1 + 73,
			Bin8,
			nil,
		},
		{
			[]byte("Testing, 1... 2... 3... This is a test of the emergency broadcast system. This is only a test. " +
				"Please do not be concerned. Everything is just fine... Uhhh, " +
				"wait... I am getting some new information now."),
			1 + 1 + 202,
			Bin8,
			nil,
		},
		{
			[]byte("Testing, 1... 2... 3... This is a test of the emergency broadcast system. This is only a test. " +
				"Please do not be concerned. Everything is just fine... Uhhh, " +
				"wait... I am getting some new information now. Okay--I don't want anyone to worry... " +
				"But it seems we have been invaded by an alien race."),
			1 + 2 + 292,
			Bin16,
			nil,
		},
		{
			[]byte("Testing, 1... 2... 3... This is a test of the emergency broadcast system. This is only a test. " +
				"Please do not be concerned. Everything is just fine... Uhhh, " +
				"wait... I am getting some new information now. Okay--I don't want anyone to worry... " +
				"But it seems we have been invaded by an alien race. " +
				"They have come to every major city and have established some kind of military approach upon meeting" +
				" some people in government. I will be sure to report back as soon as I... AHhh...... shhhwoooshhh..." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................." +
				"...................................................................................................."),
			1 + 2 + 3692,
			Bin16,
			nil,
		},
		{
			bytes.Repeat([]byte(" [We have come to take over]"), 1<<20),
			1 + 4 + 29360128,
			Bin32,
			nil,
		},
		{
			bytes.Repeat([]byte(" [We have come to take over]"), 1<<21),
			1 + 4 + 58720256,
			Bin32,
			nil,
		},
	}

	// testing bytes encoding
	for i, tt := range binTests {
		out, err := encBin(tt.str)
		if err != nil {
			t.Errorf("%s", err)
			continue
		}
		if out[0] != tt.exp {
			t.Errorf("error: type byte encoded incorrectly, got=%x, wanted=%x\n", out[0], tt.exp)
		}
		if len(out) != tt.len {
			t.Errorf("error: length of encoded output is incorrect, got=%d, wanted=%d\n", len(out), tt.len)
		}
		binTests[i].bin = out
	}

	// testing bytes decoding
	for _, tt := range binTests {
		out, err := decBin(tt.bin)
		if err != nil {
			t.Errorf("%s", err)
			continue
		}
		if !bytes.Equal(out, tt.str) {
			t.Errorf("error: decoded string did not match origional string\n")
		}
	}
}

func TestBasicTypes(t *testing.T) {
	basicTests := []struct {
		typ  byte
		val  any
		bin  []byte
		size int
		enc  func(p []byte)
		dec  func(p []byte) any
	}{
		{
			BoolTrue,
			true,
			[]byte{BoolTrue},
			1,
			func(p []byte) { encBool(p, true) },
			func(p []byte) any { return decBool(p) },
		},
		{
			BoolFalse,
			false,
			[]byte{BoolFalse},
			1,
			func(p []byte) { encBool(p, false) },
			func(p []byte) any { return decBool(p) },
		},
		{
			Nil,
			nil,
			[]byte{Nil},
			1,
			func(p []byte) { encNil(p) },
			func(p []byte) any { return decNil(p) },
		},
		{
			Uint8,
			uint8(4),
			[]byte{Uint8, 4},
			2,
			func(p []byte) { encUint8(p, 4) },
			func(p []byte) any { return decUint8(p) },
		},
		{
			Uint16,
			uint16(4444),
			[]byte{Uint16, 0x11, 0x5c},
			3,
			func(p []byte) { encUint16(p, 4444) },
			func(p []byte) any { return decUint16(p) },
		},
		{
			Uint32,
			uint32(44444444),
			[]byte{Uint32, 0x2, 0xa6, 0x2b, 0x1c},
			5,
			func(p []byte) { encUint32(p, 44444444) },
			func(p []byte) any { return decUint32(p) },
		},
		{
			Uint64,
			uint64(444444444444),
			[]byte{Uint64, 0x0, 0x0, 0x0, 0x67, 0x7a, 0xf4, 0x7, 0x1c},
			9,
			func(p []byte) { encUint64(p, 444444444444) },
			func(p []byte) any { return decUint64(p) },
		},
		{
			Int8,
			int8(5),
			[]byte{Int8, 5},
			2,
			func(p []byte) { encInt8(p, 5) },
			func(p []byte) any { return decInt8(p) },
		},
		{
			Int16,
			int16(5555),
			[]byte{Int16, 0x15, 0xb3},
			3,
			func(p []byte) { encInt16(p, 5555) },
			func(p []byte) any { return decInt16(p) },
		},
		{
			Int32,
			int32(55555555),
			[]byte{Int32, 0x3, 0x4f, 0xb5, 0xe3},
			5,
			func(p []byte) { encInt32(p, 55555555) },
			func(p []byte) any { return decInt32(p) },
		},
		{
			Int64,
			int64(555555555555),
			[]byte{Int64, 0x0, 0x0, 0x0, 0x81, 0x59, 0xb1, 0x8, 0xe3},
			9,
			func(p []byte) { encInt64(p, 555555555555) },
			func(p []byte) any { return decInt64(p) },
		},
		{
			Float32,
			float32(3.14159),
			[]byte{Float32, 0x40, 0x49, 0xf, 0xd0},
			5,
			func(p []byte) { encFloat32(p, 3.14159) },
			func(p []byte) any { return decFloat32(p) },
		},
		{
			Float64,
			float64(3.14159),
			[]byte{Float64, 0x40, 0x9, 0x21, 0xf9, 0xf0, 0x1b, 0x86, 0x6e},
			9,
			func(p []byte) { encFloat64(p, 3.14159) },
			func(p []byte) any { return decFloat64(p) },
		},
	}

	for _, tt := range basicTests {

		// create buffer of the correct size
		buf := make([]byte, tt.size)

		// run encoding function
		tt.enc(buf)

		// check to make sure encodings match
		if !bytes.Equal(buf, tt.bin) {
			t.Errorf("encodings do not match: got=%#v, wanted=%#v\n", buf, tt.bin)
		}

		// run decoding function
		out := tt.dec(buf)

		// compare results
		if out != tt.val {
			t.Errorf("decoded result does not match expected value (%#v != %#v)\n", out, tt.val)
		}
	}
}

/*
func TestStrType(t *testing.T) {

	key := "k1"

	in := Str("this is my first string")
	err := SetAs(testdb, key, &in)
	if err != nil {
		t.Errorf("SetAs failed: %s", err)
	}

	var out Str
	err = GetAs(testdb, key, &out)
	if err != nil {
		t.Errorf("GetAs failed: %s", err)
	}
	fmt.Printf("%#v\n", out)
}

func TestNumType(t *testing.T) {

	key := "k2"

	in := Num("1234.35")
	err := SetAs(testdb, key, &in)
	if err != nil {
		t.Errorf("SetAs failed: %s", err)
	}

	var out Num
	err = GetAs(testdb, key, &out)
	if err != nil {
		t.Errorf("GetAs failed: %s", err)
	}
	fmt.Printf("%#v\n", out)
}
*/
