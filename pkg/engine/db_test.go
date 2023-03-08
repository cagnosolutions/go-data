package engine

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDB(t *testing.T) {

	const path = "my/db"

	// open db
	db, err := OpenDB(path)
	if err != nil {
		t.Errorf("open: %s\n", err)
	}

	// create users
	users, err := db.Create("users.db")
	if err != nil {
		t.Errorf("create: %s\n", err)
	}

	// insert a few users
	u1 := User{1, "John Doe", "jdoe@example.com", true}
	id, err := users.Insert(&u1)
	if err != nil {
		t.Errorf("insert: %s\n", err)
	}
	fmt.Printf("inserted user with id=%d\n", id)

	u2 := User{2, "Rick Frost", "rfrost@example.com", true}
	id, err = users.Insert(&u2)
	if err != nil {
		t.Errorf("insert: %s\n", err)
	}
	fmt.Printf("inserted user with id=%d\n", id)

	u3 := User{3, "Jack Miller", "jmiller@example.com", true}
	id, err = users.Insert(&u3)
	if err != nil {
		t.Errorf("insert: %s\n", err)
	}
	fmt.Printf("inserted user with id=%d\n", id)

	err = users.Commit()
	if err != nil {
		t.Errorf("users commit: %s\n", err)
	}

	var foundUser User
	err = users.FindOne(2, &foundUser)
	if err != nil {
		t.Errorf("find one: %s\n", err)
	}
	fmt.Printf("found user %d: %+v\n", 2, foundUser)

	// close db
	err = db.Close()
	if err != nil {
		t.Errorf("close: %s\n", err)
	}

	// err = os.RemoveAll(path)
	// if err != nil {
	// 	t.Errorf("remove all: %s\n", err)
	// }
}

type User struct {
	ID       uint32 `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

func (u *User) GetID() uint32 {
	return u.ID
}

func (u *User) SetID(id uint32) {
	u.ID = id
}

func (u *User) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &u)
}
