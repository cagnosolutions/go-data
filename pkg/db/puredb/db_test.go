package puredb

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// implement `Record` interface
type User struct {
	ID     int    `json:"_id"`
	FName  string `json:"f_name"`
	LName  string `json:"l_name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

func (u *User) GetID() int {
	return u.ID
}

func (u *User) SetID(id int) {
	u.ID = id
}

// implement `RecordSet` interface
type Users []User

func (u Users) Len() int {
	return len(u)
}

var base = "dbtesting/"

func getRandomID(a []int) int {
	return a[len(a)/2]
}

func TestOpenPureDB(t *testing.T) {
	// open db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to make sure db is open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_CreateTable(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name, and full path
	name := "users.json"
	path := filepath.ToSlash(filepath.Join(base, name))
	// create table
	err = db.MakeCollection(name)
	if err != nil {
		t.Fatal(err)
	}
	// make sure path is there
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		t.Fatal("expected the file to be there and it wasn't")
	}
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_Insert(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name
	name := "users.json"
	// create three new users
	user1 := &User{
		FName:  "John",
		LName:  "Doe",
		Email:  "jdoe@gmail.com",
		Active: true,
	}
	user2 := &User{
		FName:  "Jane",
		LName:  "Doe",
		Email:  "jdoe@gmail.com",
		Active: true,
	}
	user3 := &User{
		FName:  "Rex",
		LName:  "Doe",
		Email:  "rdoe@gmail.com",
		Active: true,
	}
	user4 := &User{
		FName:  "Felix",
		LName:  "Smith",
		Email:  "fsmith@gmail.com",
		Active: true,
	}
	user5 := &User{
		FName:  "Maggie",
		LName:  "Smith",
		Email:  "msmith@gmail.com",
		Active: true,
	}
	// insert user1
	id, err := db.Insert(name, user1)
	if err != nil {
		t.Fatalf("something weird happened: %s\n", err.Error())
	}
	fmt.Printf("successfully inserted record, got id=%d\n", id)
	// insert user2
	id, err = db.Insert(name, user2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("successfully inserted record, got id=%d\n", id)
	// insert user3
	id, err = db.Insert(name, user3)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("successfully inserted record, got id=%d\n", id)
	// insert user4
	id, err = db.Insert(name, user4)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("successfully inserted record, got id=%d\n", id)
	// insert user5
	id, err = db.Insert(name, user5)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("successfully inserted record, got id=%d\n", id)
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_Update(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name
	name := "users.json"
	// create a user to update
	user := &User{
		FName:  "Bruce",
		LName:  "Wayne",
		Email:  "bwayne@batcave.com",
		Active: false,
	}
	// get table info
	ti := db.GetCollectionInfo(name)
	// get random id from table records
	recIds, _ := ti.GetRecords()
	id := getRandomID(recIds)
	fmt.Printf(">>> attempting to update record with id: %d\n", id)
	// update user
	err = db.Update(name, id, user)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("successfully updated record\n")
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_Search(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name
	name := "users.json"
	// create record to return into
	var users []User
	// return users using search (pass pointer to users)
	n, err := db.Search(name, `"_id":1;"_id":3;"_id":5`, &users)
	if err != nil {
		t.Fatalf("returning: %s\n", err)
	}
	// print user details
	fmt.Printf("returned %d users\n", n)
	for _, user := range users {
		fmt.Printf("\tuser=%+v\n", user)
	}
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_Return(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name
	name := "users.json"
	// create record to return into
	var u User
	// get table info
	ti := db.GetCollectionInfo(name)
	// get random id from table records
	recIds, _ := ti.GetRecords()
	id := getRandomID(recIds)
	fmt.Printf(">>> attempting to return record with id: %d\n", id)
	// return user (pass pointer to user)
	err = db.Return(name, id, &u)
	if err != nil {
		t.Fatalf("returning: %s\n", err)
	}
	// print user details
	fmt.Printf("user(%d): %+v\n", id, u)
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_ReturnAll(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name
	name := "users.json"
	// create record set to return into
	var users []User
	// return all users (pass pointer to users)
	n, err := db.ReturnAll(name, &users)
	if err != nil {
		t.Fatalf("returning: %s\n", err)
	}
	// print user details
	fmt.Printf("returned %d users\n", n)
	for _, user := range users {
		fmt.Printf("\tuser=%+v\n", user)
	}
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_Delete(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name
	name := "users.json"
	// get table info
	ti := db.GetCollectionInfo(name)
	// get random id from table records
	recIds, _ := ti.GetRecords()
	id := getRandomID(recIds)
	fmt.Printf(">>> attempting to delete record with id: %d\n", id)
	// delete a user
	err = db.Delete(name, id)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("successfully updated record\n")
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPureDB_DropTable(t *testing.T) {
	// open the db
	db, err := Open(base)
	if err != nil {
		t.Fatal(err)
	}
	// check to ensure it's open
	if !db.IsOpen() {
		t.Fatal("expected table to be open")
	}
	// create table name, and full path
	name := "users.json"
	path := filepath.ToSlash(filepath.Join(base, name))

	// drop table
	err = db.DropCollection(name)
	if err != nil {
		t.Fatal(err)
	}
	// check path again to make sure it's gone
	_, err = os.Stat(path)
	if os.IsExist(err) {
		t.Fatal("expected the file to be gone and it wasn't")
	}
	// close db
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCleanup(t *testing.T) {
	// clean up database folder
	err := os.Remove(base)
	if err != nil {
		t.Error(err)
	}
}
