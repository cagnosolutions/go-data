package dopedb

import (
	"bytes"
	"reflect"
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
		buf := make([]byte, tt.len)
		encStr(buf, tt.str)

		if buf[0] != tt.exp {
			t.Errorf("error: type byte encoded incorrectly, got=%x, wanted=%x\n", buf[0], tt.exp)
		}
		if len(buf) != tt.len {
			t.Errorf("error: length of encoded output is incorrect, got=%d, wanted=%d\n", len(buf), tt.len)
		}
		strTests[i].bin = buf
	}

	// testing string decoding
	for _, tt := range strTests {
		out := decStr(tt.bin)
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
		buf := make([]byte, tt.len)
		encBin(buf, tt.str)

		if buf[0] != tt.exp {
			t.Errorf("error: type byte encoded incorrectly, got=%x, wanted=%x\n", buf[0], tt.exp)
		}
		if len(buf) != tt.len {
			t.Errorf("error: length of encoded output is incorrect, got=%d, wanted=%d\n", len(buf), tt.len)
		}
		binTests[i].bin = buf
	}

	// testing bytes decoding
	for _, tt := range binTests {
		out := decBin(tt.bin)
		if !bytes.Equal(out, tt.str) {
			t.Errorf("error: decoded string did not match origional string\n")
		}
	}
}

/*
func _TestBasicTypes(t *testing.T) {
	basicTests := []struct {
		typ  byte
		val  any
		bin  []byte
		size int
		enc  func(p []byte)
		dec  func(p []byte) any
	}{
		// {
		// 	BoolTrue,
		// 	true,
		// 	[]byte{BoolTrue},
		// 	1,
		// 	func(p []byte) { WriteBool(p, true) },
		// 	func(p []byte) any { return ReadBool(p) },
		// },
		// {
		// 	BoolFalse,
		// 	false,
		// 	[]byte{BoolFalse},
		// 	1,
		// 	func(p []byte) { WriteBool(p, false) },
		// 	func(p []byte) any { return ReadBool(p) },
		// },
		// {
		// 	Nil,
		// 	nil,
		// 	[]byte{Nil},
		// 	1,
		// 	func(p []byte) { WriteNil(p) },
		// 	func(p []byte) any { return ReadNil(p) },
		// },
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
		// {
		// 	Float32,
		// 	float32(3.14159),
		// 	[]byte{Float32, 0x40, 0x49, 0xf, 0xd0},
		// 	5,
		// 	func(p []byte) { WriteFloat32(p, 3.14159) },
		// 	func(p []byte) any { return ReadFloat32(p) },
		// },
		// {
		// 	Float64,
		// 	float64(3.14159),
		// 	[]byte{Float64, 0x40, 0x9, 0x21, 0xf9, 0xf0, 0x1b, 0x86, 0x6e},
		// 	9,
		// 	func(p []byte) { WriteFloat64(p, 3.14159) },
		// 	func(p []byte) any { return decFloat64(p) },
		// },
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

func TestEncoderAndDecoderSmall(t *testing.T) {

	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)

	var f1 float32 = 3.14159
	err := enc.Encode(f1)
	if err != nil {
		t.Errorf("encoding error: %s", err)
	}
	fmt.Printf("f1: %T %#v\n", f1, f1)
	fmt.Printf("buf: %#v\n", buf.Bytes())

	dec := NewDecoder(buf)

	var f2 any
	// _, err = ReadFloat32(buf, &f2)
	// if err != nil {
	// 	return
	// }
	err = dec.Decode(&f2)
	if err != nil {
		t.Errorf("deocding error: %s", err)
	}
	fmt.Printf("f2: %T %#v\n", f2, f2)

}

func TestEncoderAndDecoder(t *testing.T) {
	basicTests := []struct {
		typ  byte
		val  any
		bin  []byte
		size int
	}{
		{
			BoolTrue,
			true,
			[]byte{BoolTrue},
			1,
		},
		{
			BoolFalse,
			false,
			[]byte{BoolFalse},
			1,
		},
		{
			Nil,
			nil,
			[]byte{Nil},
			1,
		},
		{
			Uint8,
			uint8(4),
			[]byte{Uint8, 4},
			2,
		},
		{
			Uint16,
			uint16(4444),
			[]byte{Uint16, 0x11, 0x5c},
			3,
		},
		{
			Uint32,
			uint32(44444444),
			[]byte{Uint32, 0x2, 0xa6, 0x2b, 0x1c},
			5,
		},
		{
			Uint64,
			uint64(444444444444),
			[]byte{Uint64, 0x0, 0x0, 0x0, 0x67, 0x7a, 0xf4, 0x7, 0x1c},
			9,
		},
		{
			Int8,
			int8(5),
			[]byte{Int8, 5},
			2,
		},
		{
			Int16,
			int16(5555),
			[]byte{Int16, 0x15, 0xb3},
			3,
		},
		{
			Int32,
			int32(55555555),
			[]byte{Int32, 0x3, 0x4f, 0xb5, 0xe3},
			5,
		},
		{
			Int64,
			int64(555555555555),
			[]byte{Int64, 0x0, 0x0, 0x0, 0x81, 0x59, 0xb1, 0x8, 0xe3},
			9,
		},
		{
			Float32,
			float32(3.14159),
			[]byte{Float32, 0x40, 0x49, 0xf, 0xd0},
			5,
		},
		{
			Float64,
			float64(3.14159),
			[]byte{Float64, 0x40, 0x9, 0x21, 0xf9, 0xf0, 0x1b, 0x86, 0x6e},
			9,
		},
	}

	for _, tt := range basicTests {

		// create buffer of the correct size
		buf := new(bytes.Buffer)
		// setup primitiveTypes
		enc := NewEncoder(buf)

		// run encoding function
		err := enc.Encode(tt.val)
		if err != nil {
			t.Errorf("error encoding: %s\n", err)
		}

		// check to make sure encodings match
		if !bytes.Equal(buf.Bytes(), tt.bin) {
			t.Errorf("encodings do not match: got=%#v, wanted=%#v\n", buf, tt.bin)
		}

		// run decoding function
		dec := NewDecoder(buf)
		var out any
		err = dec.Decode(&out)
		if err != nil {
			t.Errorf("error decoding: %s\n", err)
		}

		// compare results
		if out != tt.val {
			t.Errorf("decoded result does not match expected value (%#v != %#v)\n", out, tt.val)
		}

		fmt.Printf("decoded: %T %#v\n", out, out)

	}
}

func TestTypeSwitch(t *testing.T) {
	testTypes := []struct {
		val any
		exp int
	}{
		{
			true,
			BoolTrue,
		},
		{
			false,
			BoolFalse,
		},
		{
			"foobar",
			FixStr,
		},
		{
			"this is another string type that should be str8",
			Str8,
		},
		{
			[]byte("foobar"),
			Bin8,
		},
		{
			int8(123),
			Int8,
		},
		{
			int16(12345),
			Int16,
		},
		{
			int32(123456789),
			Int32,
		},
		{
			int64(1234567890123456789),
			Int64,
		},
		{
			uint8(123),
			Uint8,
		},
		{
			uint16(12345),
			Uint16,
		},
		{
			uint32(123456789),
			Uint32,
		},
		{
			uint64(1234567890123456789),
			Uint64,
		},
		{
			map[string]any{"foo": 1, "bar": 2, "baz": 3},
			FixMap,
		},
		{
			val: map[string]any{
				"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7, "h": 8, "i": 9, "j": 10, "k": 11, "l": 12,
				"m": 13,
				"n": 14, "o": 15, "p": 16, "q": 17, "r": 18, "s": 19, "t": 20, "u": 21, "v": 22, "w": 23, "x": 24,
				"y": 25, "z": 26,
			},
			exp: Map16,
		},
		{
			map[int]int{1: 1, 2: 2, 3: 3},
			FixMap, // will use reflect
		},
		{
			[]any{1, 2, 3, 4, 5, 6, 7, 8, 9},
			FixArray,
		},
		{
			[]any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			Array16,
		},
		{
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			FixArray, // will use reflect
		},
	}
	for _, tt := range testTypes {

		t1 := time.Now()
		typ := getType(tt.val)
		t2 := time.Now().Sub(t1)

		if typ != tt.exp {
			t.Errorf("error: got=%x (%s), expected=%x (%s)\n", typ, typeToString[typ], tt.exp, typeToString[tt.exp])
		}
		fmt.Printf("got type %s, took %s\n", typeToString[typ], t2)
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

func switchTest(value interface{}) int {
	switch switchType := value.(type) {
	case nil:
		if switchType == nil { // make the compiler happy.
			return 666
		}
		return 1
	case string:
		return 2
	case int:
		return 3
	case int8:
		return 4
	case int16:
		return 5
	case int32:
		return 6
	case int64:
		return 7
	case float32:
		return 8
	case float64:
		return 9
	case uint:
		return 10
	}
	return 0
}

func reflectTest(value interface{}) int {
	rvalue := reflect.ValueOf(value)
	switch rvalue.Kind() {
	case reflect.Ptr:
		return 1
	case reflect.String:
		return 2
	case reflect.Int:
		return 3
	case reflect.Int8:
		return 4
	case reflect.Int16:
		return 5
	case reflect.Int32:
		return 6
	case reflect.Int64:
		return 7
	case reflect.Float32:
		return 8
	case reflect.Float64:
		return 9
	case reflect.Uint:
		return 10
	}
	return 0
}

var globalSum int

func makeWeirdList() []interface{} {
	var list []interface{}
	for i := 0; i < 10; i++ {
		list = append(list, []interface{}{"string", 5, 5.5, nil, int8(5), uint(3), uint32(6)}...)
	}
	return list
}

func Benchmark_TypeSwitch(b *testing.B) {
	sum := 0
	list := makeWeirdList()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range list {
			sum = sum + switchTest(val)
		}
	}
	globalSum = sum
}
func Benchmark_Reflect(b *testing.B) {
	sum := 0
	list := makeWeirdList()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range list {
			sum = sum + reflectTest(val)
		}
	}
	globalSum = sum

}

// Make sure that both typeSwitch and reflection evaluate the weird list and get the same result.
func Test_Benchmark(t *testing.T) {
	list := makeWeirdList()
	sumReflect := 0
	for _, val := range list {
		sumReflect = sumReflect + reflectTest(val)
	}
	sumSwitch := 0
	for _, val := range list {
		sumSwitch = sumSwitch + reflectTest(val)
	}
	if sumReflect != sumSwitch {
		t.Errorf("reflectiong and typeswich values don't match")
	}

}
