package dopedb

import (
	"encoding/json"
	"fmt"
	"testing"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (u *User) Decode(v []byte) error {
	var usr *User
	err := json.Unmarshal(v, &usr)
	if err != nil {
		fmt.Printf("DECODE ERROR: %s", err)
	}
	*u = *usr
	return err
}

func (u *User) Encode() ([]byte, error) {
	return json.Marshal(u)
}

func TestDBGetAs(t *testing.T) {
	db, err := NewDB(nil)
	if err != nil {
		t.Errorf("open db: %s", err)
	}
	err = SetAs(db, "user1", &User{69, "Todd Packer"})
	if err != nil {
		t.Errorf("SetAs failed: %s", err)
	}
	var usr User
	err = GetAs(db, "user1", &usr)
	if err != nil {
		t.Errorf("getAs failed: %s", err)
	}
	fmt.Printf("%#v\n", usr)
}

// func TestDBParseCmd(t *testing.T) {
// 	db, err := NewDB(nil)
// 	if err != nil {
// 		t.Errorf("open db: %s", err)
// 	}
// 	cmds := []string{
// 		"set foo 1",
// 		"get foo",
// 		"incr foo",
// 		"get foo",
// 		"decr foo",
// 		"get foo",
// 		"zset list foo bar baz",
// 		"zget list",
// 		"hset user:1 id 1 name joe",
// 		"hget user:1",
// 		"hset user:1 id 1 name",
// 		"hget user:1",
// 		"get user:1",
// 	}
//
// 	for _, cmd := range cmds {
// 		got, err := parseCmd(db, cmd)
// 		if err != nil {
// 			t.Errorf("error: %s", err)
// 		}
// 		fmt.Printf("cmd: %q\ngot: %q\n\n", cmd, got)
// 	}
// }
