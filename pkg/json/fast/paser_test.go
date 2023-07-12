package fast

import (
	"bytes"
	"fmt"
	"testing"
)

type GetTest struct {
	desc string
	json string
	path []string

	isErr   bool
	isFound bool

	data interface{}
}

type SetTest struct {
	desc    string
	json    string
	setData string
	path    []string

	isErr   bool
	isFound bool

	data interface{}
}

type DeleteTest struct {
	desc string
	json string
	path []string

	data interface{}
}

var deleteTests = []DeleteTest{
	{
		desc: "Delete test key",
		json: `{"test":"input"}`,
		path: []string{"test"},
		data: `{}`,
	},
	{
		desc: "Delete object",
		json: `{"test":"input"}`,
		path: []string{},
		data: ``,
	},
	{
		desc: "Delete a nested object",
		json: `{"test":"input","new.field":{"key": "new object"}}`,
		path: []string{"new.field", "key"},
		data: `{"test":"input","new.field":{}}`,
	},
	{
		desc: "Deleting a key that doesn't exist should return the same object",
		json: `{"test":"input"}`,
		path: []string{"test2"},
		data: `{"test":"input"}`,
	},
	{
		desc: "Delete object in an array",
		json: `{"test":[{"key":"val-obj1"}]}`,
		path: []string{"test", "[0]"},
		data: `{"test":[]}`,
	},
	{
		desc: "Deleting a object in an array that doesn't exists should return the same object",
		json: `{"test":[{"key":"val-obj1"}]}`,
		path: []string{"test", "[1]"},
		data: `{"test":[{"key":"val-obj1"}]}`,
	},
	{
		desc: "Delete a complex object in a nested array",
		json: `{"test":[{"key":[{"innerKey":"innerKeyValue"}]}]}`,
		path: []string{"test", "[0]", "key", "[0]"},
		data: `{"test":[{"key":[]}]}`,
	},
	{
		desc: "Delete known key (simple type within nested array)",
		json: `{"test":[{"key":["innerKey"]}]}`,
		path: []string{"test", "[0]", "key", "[0]"},
		data: `{"test":[{"key":[]}]}`,
	},
	{
		desc: "Delete in empty json",
		json: `{}`,
		path: []string{},
		data: ``,
	},
	{
		desc: "Delete empty array",
		json: `[]`,
		path: []string{},
		data: ``,
	},
	{
		desc: "Deleting non json should return the same value",
		json: `1.323`,
		path: []string{"foo"},
		data: `1.323`,
	},
	{
		desc: "Delete known key (top level array)",
		json: `[{"key":"val-obj1"}]`,
		path: []string{"[0]"},
		data: `[]`,
	},
	{
		// This test deletes the key instead of returning a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc: `malformed with trailing whitespace`,
		json: `{"a":1 `,
		path: []string{"a"},
		data: `{ `,
	},
	{
		// This test dels the key instead of returning a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc: "malformed 'colon chain', delete b",
		json: `{"a":"b":"c"}`,
		path: []string{"b"},
		data: `{"a":}`,
	},
	{
		desc: "Delete object without inner array",
		json: `{"a": {"b": 1}, "b": 2}`,
		path: []string{"b"},
		data: `{"a": {"b": 1}}`,
	},
	{
		desc: "Delete object without inner array",
		json: `{"a": [{"b": 1}], "b": 2}`,
		path: []string{"b"},
		data: `{"a": [{"b": 1}]}`,
	},
	{
		desc: "Delete object without inner array",
		json: `{"a": {"c": {"b": 3}, "b": 1}, "b": 2}`,
		path: []string{"a", "b"},
		data: `{"a": {"c": {"b": 3}}, "b": 2}`,
	},
	{
		desc: "Delete object without inner array",
		json: `{"a": [{"c": {"b": 3}, "b": 1}], "b": 2}`,
		path: []string{"a", "[0]", "b"},
		data: `{"a": [{"c": {"b": 3}}], "b": 2}`,
	},
	{
		desc: "Remove trailing comma if last object is deleted",
		json: `{"a": "1", "b": "2"}`,
		path: []string{"b"},
		data: `{"a": "1"}`,
	},
	{
		desc: "Correctly delete first element with space-comma",
		json: `{"a": "1" ,"b": "2" }`,
		path: []string{"a"},
		data: `{"b": "2" }`,
	},
	{
		desc: "Correctly delete middle element with space-comma",
		json: `{"a": "1" ,"b": "2" , "c": 3}`,
		path: []string{"b"},
		data: `{"a": "1" , "c": 3}`,
	},
	{
		desc: "Delete non-last key",
		json: `{"test":"input","test1":"input1"}`,
		path: []string{"test"},
		data: `{"test1":"input1"}`,
	},
	{
		desc: "Delete non-exist key",
		json: `{"test:":"input"}`,
		path: []string{"test", "test1"},
		data: `{"test:":"input"}`,
	},
	{
		desc: "Delete non-last object in an array",
		json: `[{"key":"val-obj1"},{"key2":"val-obj2"}]`,
		path: []string{"[0]"},
		data: `[{"key2":"val-obj2"}]`,
	},
	{
		desc: "Delete non-first object in an array",
		json: `[{"key":"val-obj1"},{"key2":"val-obj2"}]`,
		path: []string{"[1]"},
		data: `[{"key":"val-obj1"}]`,
	},
	{
		desc: "Issue #188: infinite loop in Delete",
		json: `^_ï¿½^C^A^@[`,
		path: []string{""},
		data: `^_ï¿½^C^A^@[`,
	},
	{
		desc: "Issue #188: infinite loop in Delete",
		json: `^_ï¿½^C^A^@{`,
		path: []string{""},
		data: `^_ï¿½^C^A^@{`,
	},
	{
		desc: "Issue #150: leading space",
		json: `   {"test":"input"}`,
		path: []string{"test"},
		data: `   {}`,
	},
}

var setTests = []SetTest{
	{
		desc:    "set unknown key (string)",
		json:    `{"test":"input"}`,
		isFound: true,
		path:    []string{"new.field"},
		setData: `"new value"`,
		data:    `{"test":"input","new.field":"new value"}`,
	},
	{
		desc:    "set known key (string)",
		json:    `{"test":"input"}`,
		isFound: true,
		path:    []string{"test"},
		setData: `"new value"`,
		data:    `{"test":"new value"}`,
	},
	{
		desc:    "set unknown key (object)",
		json:    `{"test":"input"}`,
		isFound: true,
		path:    []string{"new.field"},
		setData: `{"key": "new object"}`,
		data:    `{"test":"input","new.field":{"key": "new object"}}`,
	},
	{
		desc:    "set known key (object)",
		json:    `{"test":"input"}`,
		isFound: true,
		path:    []string{"test"},
		setData: `{"key": "new object"}`,
		data:    `{"test":{"key": "new object"}}`,
	},
	{
		desc:    "set known key (object within array)",
		json:    `{"test":[{"key":"val-obj1"}]}`,
		isFound: true,
		path:    []string{"test", "[0]"},
		setData: `{"key":"new object"}`,
		data:    `{"test":[{"key":"new object"}]}`,
	},
	{
		desc:    "set unknown key (replace object)",
		json:    `{"test":[{"key":"val-obj1"}]}`,
		isFound: true,
		path:    []string{"test", "newKey"},
		setData: `"new object"`,
		data:    `{"test":{"newKey":"new object"}}`,
	},
	{
		desc:    "set unknown key (complex object within nested array)",
		json:    `{"test":[{"key":[{"innerKey":"innerKeyValue"}]}]}`,
		isFound: true,
		path:    []string{"test", "[0]", "key", "[0]", "newInnerKey"},
		setData: `{"key":"new object"}`,
		data:    `{"test":[{"key":[{"innerKey":"innerKeyValue","newInnerKey":{"key":"new object"}}]}]}`,
	},
	{
		desc:    "set known key (complex object within nested array)",
		json:    `{"test":[{"key":[{"innerKey":"innerKeyValue"}]}]}`,
		isFound: true,
		path:    []string{"test", "[0]", "key", "[0]", "innerKey"},
		setData: `{"key":"new object"}`,
		data:    `{"test":[{"key":[{"innerKey":{"key":"new object"}}]}]}`,
	},
	{
		desc:    "set unknown key (object, partial subtree exists)",
		json:    `{"test":{"input":"output"}}`,
		isFound: true,
		path:    []string{"test", "new.field"},
		setData: `{"key":"new object"}`,
		data:    `{"test":{"input":"output","new.field":{"key":"new object"}}}`,
	},
	{
		desc:    "set unknown key (object, empty partial subtree exists)",
		json:    `{"test":{}}`,
		isFound: true,
		path:    []string{"test", "new.field"},
		setData: `{"key":"new object"}`,
		data:    `{"test":{"new.field":{"key":"new object"}}}`,
	},
	{
		desc:    "set unknown key (object, no subtree exists)",
		json:    `{"test":"input"}`,
		isFound: true,
		path:    []string{"new.field", "nested", "value"},
		setData: `{"key": "new object"}`,
		data:    `{"test":"input","new.field":{"nested":{"value":{"key": "new object"}}}}`,
	},
	{
		desc:    "set in empty json",
		json:    `{}`,
		isFound: true,
		path:    []string{"foo"},
		setData: `"null"`,
		data:    `{"foo":"null"}`,
	},
	{
		desc:    "set subtree in empty json",
		json:    `{}`,
		isFound: true,
		path:    []string{"foo", "bar"},
		setData: `"null"`,
		data:    `{"foo":{"bar":"null"}}`,
	},
	{
		desc:    "set in empty string - not found",
		json:    ``,
		isFound: false,
		path:    []string{"foo"},
		setData: `"null"`,
		data:    ``,
	},
	{
		desc:    "set in Number - not found",
		json:    `1.323`,
		isFound: false,
		path:    []string{"foo"},
		setData: `"null"`,
		data:    `1.323`,
	},
	{
		desc:    "set known key (top level array)",
		json:    `[{"key":"val-obj1"}]`,
		isFound: true,
		path:    []string{"[0]", "key"},
		setData: `"new object"`,
		data:    `[{"key":"new object"}]`,
	},
	{
		desc:    "set unknown key (trailing whitespace)",
		json:    `{"key":"val-obj1"}  `,
		isFound: true,
		path:    []string{"alt-key"},
		setData: `"new object"`,
		data:    `{"key":"val-obj1","alt-key":"new object"}  `,
	},
	{
		// This test sets the key instead of returning a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc:    `malformed with trailing whitespace`,
		json:    `{"a":1 `,
		path:    []string{"a"},
		setData: `2`,
		isFound: true,
		data:    `{"a":2 `,
	},
	{
		// This test sets the key instead of returning a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc:    "malformed 'colon chain', set second string",
		json:    `{"a":"b":"c"}`,
		path:    []string{"b"},
		setData: `"d"`,
		isFound: true,
		data:    `{"a":"b":"d"}`,
	},
	{
		desc:    "set indexed path to object on empty JSON",
		json:    `{}`,
		path:    []string{"top", "[0]", "middle", "[0]", "bottom"},
		setData: `"value"`,
		isFound: true,
		data:    `{"top":[{"middle":[{"bottom":"value"}]}]}`,
	},
	{
		desc:    "set indexed path on existing object with object",
		json:    `{"top":[{"middle":[]}]}`,
		path:    []string{"top", "[0]", "middle", "[0]", "bottom"},
		setData: `"value"`,
		isFound: true,
		data:    `{"top":[{"middle":[{"bottom":"value"}]}]}`,
	},
	{
		desc:    "set indexed path on existing object with value",
		json:    `{"top":[{"middle":[]}]}`,
		path:    []string{"top", "[0]", "middle", "[0]"},
		setData: `"value"`,
		isFound: true,
		data:    `{"top":[{"middle":["value"]}]}`,
	},
	{
		desc:    "set indexed path on empty object with value",
		json:    `{}`,
		path:    []string{"top", "[0]", "middle", "[0]"},
		setData: `"value"`,
		isFound: true,
		data:    `{"top":[{"middle":["value"]}]}`,
	},
	{
		desc:    "set indexed path on object with existing array",
		json:    `{"top":["one", "two", "three"]}`,
		path:    []string{"top", "[2]"},
		setData: `"value"`,
		isFound: true,
		data:    `{"top":["one", "two", "value"]}`,
	},
	{
		desc:    "set non-exist key",
		json:    `{"test":"input"}`,
		setData: `"new value"`,
		isFound: false,
	},
	{
		desc:    "set key in invalid json",
		json:    `{"test"::"input"}`,
		path:    []string{"test"},
		setData: "new value",
		isErr:   true,
	},
	{
		desc:    "set unknown key (simple object within nested array)",
		json:    `{"test":{"key":[{"innerKey":"innerKeyValue", "innerKey2":"innerKeyValue2"}]}}`,
		isFound: true,
		path:    []string{"test", "key", "[1]", "newInnerKey"},
		setData: `"new object"`,
		data:    `{"test":{"key":[{"innerKey":"innerKeyValue", "innerKey2":"innerKeyValue2"},{"newInnerKey":"new object"}]}}`,
	},
}

var getTests = []GetTest{
	// Trivial tests
	{
		desc:    "read string",
		json:    `""`,
		isFound: true,
		data:    ``,
	},
	{
		desc:    "read number",
		json:    `0`,
		isFound: true,
		data:    `0`,
	},
	{
		desc:    "read object",
		json:    `{}`,
		isFound: true,
		data:    `{}`,
	},
	{
		desc:    "read array",
		json:    `[]`,
		isFound: true,
		data:    `[]`,
	},
	{
		desc:    "read boolean",
		json:    `true`,
		isFound: true,
		data:    `true`,
	},

	// Found key tests
	{
		desc:    "handling multiple nested keys with same name",
		json:    `{"a":[{"b":1},{"b":2},3],"c":{"c":[1,2]}} }`,
		path:    []string{"c", "c"},
		isFound: true,
		data:    `[1,2]`,
	},
	{
		desc:    "read basic key",
		json:    `{"a":"b"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:    "read basic key with space",
		json:    `{"a": "b"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:    "read composite key",
		json:    `{"a": { "b":{"c":"d" }}}`,
		path:    []string{"a", "b", "c"},
		isFound: true,
		data:    `d`,
	},
	{
		desc:    `read numberic value as string`,
		json:    `{"a": "b", "c": 1}`,
		path:    []string{"c"},
		isFound: true,
		data:    `1`,
	},
	{
		desc:    `handle multiple nested keys with same name`,
		json:    `{"a":[{"b":1},{"b":2},3],"c":{"c":[1,2]}} }`,
		path:    []string{"c", "c"},
		isFound: true,
		data:    `[1,2]`,
	},
	{
		desc:    `read string values with quotes`,
		json:    `{"a": "string\"with\"quotes"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `string\"with\"quotes`,
	},
	{
		desc:    `read object`,
		json:    `{"a": { "b":{"c":"d" }}}`,
		path:    []string{"a", "b"},
		isFound: true,
		data:    `{"c":"d" }`,
	},
	{
		desc:    `empty path`,
		json:    `{"c":"d" }`,
		path:    []string{},
		isFound: true,
		data:    `{"c":"d" }`,
	},
	{
		desc:    `formatted JSON value`,
		json:    "{\n  \"a\": \"b\"\n}",
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:    `formatted JSON value 2`,
		json:    "{\n  \"a\":\n    {\n\"b\":\n   {\"c\":\"d\",\n\"e\": \"f\"}\n}\n}",
		path:    []string{"a", "b"},
		isFound: true,
		data:    "{\"c\":\"d\",\n\"e\": \"f\"}",
	},
	{
		desc:    `whitespace`,
		json:    " \n\r\t{ \n\r\t\"whitespace\" \n\r\t: \n\r\t333 \n\r\t} \n\r\t",
		path:    []string{"whitespace"},
		isFound: true,
		data:    "333",
	},
	{
		desc:    `escaped backslash quote`,
		json:    `{"a": "\\\""}`,
		path:    []string{"a"},
		isFound: true,
		data:    `\\\"`,
	},
	{
		desc:    `unescaped backslash quote`,
		json:    `{"a": "\\"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `\\`,
	},
	{
		desc:    `unicode in JSON`,
		json:    `{"a": "15°C"}`,
		path:    []string{"a"},
		isFound: true,
		data:    `15°C`,
	},
	{
		desc:    `no padding + nested`,
		json:    `{"a":{"a":"1"},"b":2}`,
		path:    []string{"b"},
		isFound: true,
		data:    `2`,
	},
	{
		desc:    `no padding + nested + array`,
		json:    `{"a":{"b":[1,2]},"c":3}`,
		path:    []string{"c"},
		isFound: true,
		data:    `3`,
	},
	{
		desc:    `empty key`,
		json:    `{"":{"":{"":true}}}`,
		path:    []string{"", "", ""},
		isFound: true,
		data:    `true`,
	},

	// Escaped key tests
	{
		desc:    `key with simple escape`,
		json:    `{"a\\b":1}`,
		path:    []string{"a\\b"},
		isFound: true,
		data:    `1`,
	},
	{
		desc:    `key and value with whitespace escapes`,
		json:    `{"key\b\f\n\r\tkey":"value\b\f\n\r\tvalue"}`,
		path:    []string{"key\b\f\n\r\tkey"},
		isFound: true,
		data:    `value\b\f\n\r\tvalue`, // value is not unescaped since this is Get(), but the key should work correctly
	},
	{
		desc:    `key with Unicode escape`,
		json:    `{"a\u00B0b":1}`,
		path:    []string{"a\u00B0b"},
		isFound: true,
		data:    `1`,
	},
	{
		desc:    `key with complex escape`,
		json:    `{"a\uD83D\uDE03b":1}`,
		path:    []string{"a\U0001F603b"},
		isFound: true,
		data:    `1`,
	},

	{
		// This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance
		desc:    `malformed with trailing whitespace`,
		json:    `{"a":1 `,
		path:    []string{"a"},
		isFound: true,
		data:    `1`,
	},
	{
		// This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance
		desc:    `malformed with wrong closing bracket`,
		json:    `{"a":1]`,
		path:    []string{"a"},
		isFound: true,
		data:    `1`,
	},

	// Not found key tests
	{
		desc:    `empty input`,
		json:    ``,
		path:    []string{"a"},
		isFound: false,
	},
	{
		desc:    "non-existent key 1",
		json:    `{"a":"b"}`,
		path:    []string{"c"},
		isFound: false,
	},
	{
		desc:    "non-existent key 2",
		json:    `{"a":"b"}`,
		path:    []string{"b"},
		isFound: false,
	},
	{
		desc:    "non-existent key 3",
		json:    `{"aa":"b"}`,
		path:    []string{"a"},
		isFound: false,
	},
	{
		desc:    "apply scope of parent when search for nested key",
		json:    `{"a": { "b": 1}, "c": 2 }`,
		path:    []string{"a", "b", "c"},
		isFound: false,
	},
	{
		desc:    `apply scope to key level`,
		json:    `{"a": { "b": 1}, "c": 2 }`,
		path:    []string{"b"},
		isFound: false,
	},
	{
		desc:    `handle escaped quote in key name in JSON`,
		json:    `{"key\"key": 1}`,
		path:    []string{"key"},
		isFound: false,
	},
	{
		desc:    "handling multiple keys with different name",
		json:    `{"a":{"a":1},"b":{"a":3,"c":[1,2]}}`,
		path:    []string{"a", "c"},
		isFound: false,
	},
	{
		desc:    "handling nested json",
		json:    `{"a":{"b":{"c":1},"d":4}}`,
		path:    []string{"a", "d"},
		isFound: true,
		data:    `4`,
	},
	{
		// Issue #148
		desc:    `missing key in different key same level`,
		json:    `{"s":"s","ic":2,"r":{"o":"invalid"}}`,
		path:    []string{"ic", "o"},
		isFound: false,
	},

	// Error/invalid tests
	{
		desc:    `handle escaped quote in key name in JSON`,
		json:    `{"key\"key": 1}`,
		path:    []string{"key"},
		isFound: false,
	},
	{
		desc:    `missing closing brace, but can still find key`,
		json:    `{"a":"b"`,
		path:    []string{"a"},
		isFound: true,
		data:    `b`,
	},
	{
		desc:  `missing value closing quote`,
		json:  `{"a":"b`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value closing curly brace`,
		json:  `{"a": { "b": "c"`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value closing square bracket`,
		json:  `{"a": [1, 2, 3 }`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value 1`,
		json:  `{"a":`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value 2`,
		json:  `{"a": `,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `missing value 3`,
		json:  `{"a":}`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:    `malformed array (no closing brace)`,
		json:    `{"a":[, "b":123}`,
		path:    []string{"b"},
		isFound: false,
	},
	{
		// Issue #81
		desc:    `missing key in object in array`,
		json:    `{"p":{"a":[{"u":"abc","t":"th"}]}}`,
		path:    []string{"p", "a", "[0]", "x"},
		isFound: false,
	},
	{
		// Issue #81 counter test
		desc:    `existing key in object in array`,
		json:    `{"p":{"a":[{"u":"abc","t":"th"}]}}`,
		path:    []string{"p", "a", "[0]", "u"},
		isFound: true,
		data:    "abc",
	},
	{
		// This test returns not found instead of a parse error, as checking for the malformed JSON would reduce performance
		desc:    "malformed key (followed by comma followed by colon)",
		json:    `{"a",:1}`,
		path:    []string{"a"},
		isFound: false,
	},
	{
		// This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc:    "malformed 'colon chain', lookup first string",
		json:    `{"a":"b":"c"}`,
		path:    []string{"a"},
		isFound: true,
		data:    "b",
	},
	{
		// This test returns a match instead of a parse error, as checking for the malformed JSON would reduce performance (this is not ideal)
		desc:    "malformed 'colon chain', lookup second string",
		json:    `{"a":"b":"c"}`,
		path:    []string{"b"},
		isFound: true,
		data:    "c",
	},
	// Array index paths
	{
		desc:    "last key in path is index",
		json:    `{"a":[{"b":1},{"b":"2"}, 3],"c":{"c":[1,2]}}`,
		path:    []string{"a", "[1]"},
		isFound: true,
		data:    `{"b":"2"}`,
	},
	{
		desc:    "get string from array",
		json:    `{"a":[{"b":1},"foo", 3],"c":{"c":[1,2]}}`,
		path:    []string{"a", "[1]"},
		isFound: true,
		data:    "foo",
	},
	{
		desc:    "key in path is index",
		json:    `{"a":[{"b":"1"},{"b":"2"},3],"c":{"c":[1,2]}}`,
		path:    []string{"a", "[0]", "b"},
		isFound: true,
		data:    `1`,
	},
	{
		desc: "last key in path is an index to value in array (formatted json)",
		json: `{
		    "a": [
			{
			    "b": 1
			},
			{"b":"2"},
			3
		    ],
		    "c": {
			"c": [
			    1,
			    2
			]
		    }
		}`,
		path:    []string{"a", "[1]"},
		isFound: true,
		data:    `{"b":"2"}`,
	},
	{
		desc: "key in path is index (formatted json)",
		json: `{
		    "a": [
			{"b": 1},
			{"b": "2"},
			3
		    ],
		    "c": {
			"c": [
			    1,
			    2
			]
		    }
		}`,
		path:    []string{"a", "[0]", "b"},
		isFound: true,
		data:    `1`,
	},
	{
		// Issue #178: Crash in searchKeys
		desc:    `invalid json`,
		json:    `{{{"":`,
		path:    []string{"a", "b"},
		isFound: false,
	},
	{
		desc:    `opening brace instead of closing and without key`,
		json:    `{"a":1{`,
		path:    []string{"b"},
		isFound: false,
	},
}

var getIntTests = []GetTest{
	{
		desc:    `read numeric value as number`,
		json:    `{"a": "b", "c": 1}`,
		path:    []string{"c"},
		isFound: true,
		data:    int64(1),
	},
	{
		desc:    `read numeric value as number in formatted JSON`,
		json:    "{\"a\": \"b\", \"c\": 1 \n}",
		path:    []string{"c"},
		isFound: true,
		data:    int64(1),
	},
	{
		// Issue #138: overflow detection
		desc:  `Fails because of overflow`,
		json:  `{"p":9223372036854775808}`,
		path:  []string{"p"},
		isErr: true,
	},
	{
		// Issue #138: overflow detection
		desc:  `Fails because of underflow`,
		json:  `{"p":-9223372036854775809}`,
		path:  []string{"p"},
		isErr: true,
	},
	{
		desc:  `read non-numeric value as integer`,
		json:  `{"a": "b", "c": "d"}`,
		path:  []string{"c"},
		isErr: true,
	},
	{
		desc:  `null test`,
		json:  `{"a": "b", "c": null}`,
		path:  []string{"c"},
		isErr: true,
	},
}

var getFloatTests = []GetTest{
	{
		desc:    `read numeric value as number`,
		json:    `{"a": "b", "c": 1.123}`,
		path:    []string{"c"},
		isFound: true,
		data:    float64(1.123),
	},
	{
		desc:    `read numeric value as number in formatted JSON`,
		json:    "{\"a\": \"b\", \"c\": 23.41323 \n}",
		path:    []string{"c"},
		isFound: true,
		data:    float64(23.41323),
	},
	{
		desc:  `read non-numeric value as float`,
		json:  `{"a": "b", "c": "d"}`,
		path:  []string{"c"},
		isErr: true,
	},
	{
		desc:  `null test`,
		json:  `{"a": "b", "c": null}`,
		path:  []string{"c"},
		isErr: true,
	},
}

var getStringTests = []GetTest{
	{
		desc:    `Translate Unicode symbols`,
		json:    `{"c": "test"}`,
		path:    []string{"c"},
		isFound: true,
		data:    `test`,
	},
	{
		desc:    `Translate Unicode symbols`,
		json:    `{"c": "15\u00b0C"}`,
		path:    []string{"c"},
		isFound: true,
		data:    `15°C`,
	},
	{
		desc:    `Translate supplementary Unicode symbols`,
		json:    `{"c": "\uD83D\uDE03"}`, // Smiley face (UTF16 surrogate pair)
		path:    []string{"c"},
		isFound: true,
		data:    "\U0001F603", // Smiley face
	},
	{
		desc:    `Translate escape symbols`,
		json:    `{"c": "\\\""}`,
		path:    []string{"c"},
		isFound: true,
		data:    `\"`,
	},
	{
		desc:    `key and value with whitespace escapes`,
		json:    `{"key\b\f\n\r\tkey":"value\b\f\n\r\tvalue"}`,
		path:    []string{"key\b\f\n\r\tkey"},
		isFound: true,
		data:    "value\b\f\n\r\tvalue", // value is unescaped since this is GetString()
	},
	{
		// This test checks we avoid an infinite loop for certain malformed JSON. We don't check for all malformed JSON as it would reduce performance.
		desc:    `malformed with double quotes`,
		json:    `{"a"":1}`,
		path:    []string{"a"},
		isFound: false,
		data:    ``,
	},
	{
		// More malformed JSON testing, to be sure we avoid an infinite loop.
		desc:    `malformed with double quotes, and path does not exist`,
		json:    `{"z":123,"y":{"x":7,"w":0},"v":{"u":"t","s":"r","q":0,"p":1558051800},"a":"b","c":"2016-11-02T20:10:11Z","d":"e","f":"g","h":{"i":"j""},"k":{"l":"m"}}`,
		path:    []string{"o"},
		isFound: false,
		data:    ``,
	},
	{
		desc:  `read non-string as string`,
		json:  `{"c": true}`,
		path:  []string{"c"},
		isErr: true,
	},
	{
		desc:    `empty array index`,
		json:    `[""]`,
		path:    []string{"[]"},
		isFound: false,
	},
	{
		desc:    `malformed array index`,
		json:    `[""]`,
		path:    []string{"["},
		isFound: false,
	},
	{
		desc:  `null test`,
		json:  `{"c": null}`,
		path:  []string{"c"},
		isErr: true,
	},
}

var getUnsafeStringTests = []GetTest{
	{
		desc:    `Do not translate Unicode symbols`,
		json:    `{"c": "test"}`,
		path:    []string{"c"},
		isFound: true,
		data:    `test`,
	},
	{
		desc:    `Do not translate Unicode symbols`,
		json:    `{"c": "15\u00b0C"}`,
		path:    []string{"c"},
		isFound: true,
		data:    `15\u00b0C`,
	},
	{
		desc:    `Do not translate supplementary Unicode symbols`,
		json:    `{"c": "\uD83D\uDE03"}`, // Smiley face (UTF16 surrogate pair)
		path:    []string{"c"},
		isFound: true,
		data:    `\uD83D\uDE03`, // Smiley face
	},
	{
		desc:    `Do not translate escape symbols`,
		json:    `{"c": "\\\""}`,
		path:    []string{"c"},
		isFound: true,
		data:    `\\\"`,
	},
}

var getBoolTests = []GetTest{
	{
		desc:    `read boolean true as boolean`,
		json:    `{"a": "b", "c": true}`,
		path:    []string{"c"},
		isFound: true,
		data:    true,
	},
	{
		desc:    `boolean true in formatted JSON`,
		json:    "{\"a\": \"b\", \"c\": true \n}",
		path:    []string{"c"},
		isFound: true,
		data:    true,
	},
	{
		desc:    `read boolean false as boolean`,
		json:    `{"a": "b", "c": false}`,
		path:    []string{"c"},
		isFound: true,
		data:    false,
	},
	{
		desc:    `boolean true in formatted JSON`,
		json:    "{\"a\": \"b\", \"c\": false \n}",
		path:    []string{"c"},
		isFound: true,
		data:    false,
	},
	{
		desc:  `read fake boolean true`,
		json:  `{"a": txyz}`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:  `read fake boolean false`,
		json:  `{"a": fwxyz}`,
		path:  []string{"a"},
		isErr: true,
	},
	{
		desc:    `read boolean true with whitespace and another key`,
		json:    "{\r\t\n \"a\"\r\t\n :\r\t\n true\r\t\n ,\r\t\n \"b\": 1}",
		path:    []string{"a"},
		isFound: true,
		data:    true,
	},
	{
		desc:    `null test`,
		json:    `{"a": "b", "c": null}`,
		path:    []string{"c"},
		isFound: false,
		isErr:   true,
	},
}

var getArrayTests = []GetTest{
	{
		desc:    `read array of simple values`,
		json:    `{"a": { "b":[1,2,3,4]}}`,
		path:    []string{"a", "b"},
		isFound: true,
		data:    []string{`1`, `2`, `3`, `4`},
	},
	{
		desc:    `read array via empty path`,
		json:    `[1,2,3,4]`,
		path:    []string{},
		isFound: true,
		data:    []string{`1`, `2`, `3`, `4`},
	},
	{
		desc:    `read array of objects`,
		json:    `{"a": { "b":[{"x":1},{"x":2},{"x":3},{"x":4}]}}`,
		path:    []string{"a", "b"},
		isFound: true,
		data:    []string{`{"x":1}`, `{"x":2}`, `{"x":3}`, `{"x":4}`},
	},
	{
		desc:    `read nested array`,
		json:    `{"a": [[[1]],[[2]]]}`,
		path:    []string{"a"},
		isFound: true,
		data:    []string{`[[1]]`, `[[2]]`},
	},
}

// checkFoundAndNoError checks the dataType and error return from Get*() against the test case expectations.
// Returns true the test should proceed to checking the actual data returned from Get*(), or false if the test is finished.
func getTestCheckFoundAndNoError(
	t *testing.T, testKind string, test GetTest, jtype ValueType, value interface{}, err error,
) bool {
	isFound := (err != ErrKeyNotFound)
	isErr := (err != nil && err != ErrKeyNotFound)

	if test.isErr != isErr {
		// If the call didn't match the error expectation, fail
		t.Errorf(
			"%s test '%s' isErr mismatch: expected %t, obtained %t (err %v). Value: %v", testKind, test.desc,
			test.isErr, isErr, err, value,
		)
		return false
	} else if isErr {
		// Else, if there was an error, don't fail and don't check isFound or the value
		return false
	} else if test.isFound != isFound {
		// Else, if the call didn't match the is-found expectation, fail
		t.Errorf("%s test '%s' isFound mismatch: expected %t, obtained %t", testKind, test.desc, test.isFound, isFound)
		return false
	} else if !isFound {
		// Else, if no value was found, don't fail and don't check the value
		return false
	} else {
		// Else, there was no error and a value was found, so check the value
		return true
	}
}

func TestGet(t *testing.T) {
	testKind := "get()"
	for _, tt := range getTests {
		fmt.Printf("Running: %s\n", tt.desc)
		value, dataType, _, err := get([]byte(tt.json), tt.path...)
		if getTestCheckFoundAndNoError(t, testKind, tt, dataType, value, err) {
			if tt.data == nil {
				t.Errorf("Malformed test: %v", tt)
				continue
			}

			expected := []byte(tt.data.(string))
			if !bytes.Equal(expected, value) {
				t.Errorf(
					"%s test '%s' expected to return value %v, but did returned %v instead",
					testKind, tt.desc, expected, value,
				)
			}
		}
	}
}
