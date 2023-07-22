package enigma_db

import (
	"fmt"
	"testing"
)

type User struct {
	ID       int
	Name     string
	IsActive bool
}

func TestRecordSet_AddAndGetRecord(t *testing.T) {

	// create a record set
	rs := NewRecordSet[User]("users")

	// add some records
	rs.AddRecord("user-003", User{3, "Jon Doe", true})
	rs.AddRecord("user-001", User{1, "Jack Smith", true})
	rs.AddRecord("user-007", User{7, "Dan Martin", true})
	rs.AddRecord("user-005", User{5, "Ben Rank", true})
	rs.AddRecord("user-002", User{2, "Allison Smitherspoon", true})
	rs.AddRecord("user-006", User{6, "Jackie Gringle", true})
	rs.AddRecord("user-004", User{4, "Matt Spanks", true})
	rs.AddRecord("user-009", User{9, "The Dude", true})
	rs.AddRecord("user-008", User{8, "Yoda", true})

	// get record
	user := rs.GetRecord("user-006")
	fmt.Println(user)
}
