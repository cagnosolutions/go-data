package sqlite

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	stmtShowTables = `SELECT 'tbl_name' FROM sqlite_master WHERE 'type'='table' AND 'tbl_name' NOT LIKE 'sqlite_%';`
)

type SQLiteDataSource struct {
	db *sql.DB
}

func NewSQLDataSource(db *sql.DB) *SQLiteDataSource {
	return &SQLiteDataSource{
		db: db,
	}
}

func (ds *SQLiteDataSource) ShowTables() {
	// prepare to protect against sql-injection
	stmt, err := ds.db.Prepare(stmtShowTables)
	if err != nil {
		log.Fatalf("preparing(%q): %s", stmtShowTables, err)
	}
	// execute
	_, err = stmt.Exec()
	if err != nil {
		log.Fatalf("executing prepared statment: %s", err)
	}
}

func (ds *SQLiteDataSource) CreateTable(table *Table) {
	// sql statement
	stmtCreate := table.Create()
	// prepare to protect against sql-injection
	stmt, err := ds.db.Prepare(stmtCreate)
	if err != nil {
		log.Fatalf("preparing(%q): %s", stmtCreate, err)
	}
	// execute
	_, err = stmt.Exec()
	if err != nil {
		log.Fatalf("executing prepared statment: %s", err)
	}
}

func (ds *SQLiteDataSource) DropTable(table *Table) {
	// sql statement
	dropTable := `DROP TABLE IF EXISTS user;`
	// prepare to protect against sql-injection
	stmt, err := ds.db.Prepare(dropTable)
	if err != nil {
		log.Fatalf("preparing(%q): %s", dropTable, err)
	}
	// execute
	_, err = stmt.Exec()
	if err != nil {
		log.Fatalf("executing prepared statment: %s", err)
	}
}

func (ds *SQLiteDataSource) SelectAll(table string, res []interface{}) {
	// sql statement
	findAllUsers := `SELECT * FROM %s ORDER BY id;`
	// execute select query
	rows, err := ds.db.Query(fmt.Sprintf(findAllUsers, table))
	if err != nil {
		log.Fatalf("query(%q): %s", findAllUsers, err)
	}
	defer rows.Close()
	// load data into users ptr
	for rows.Next() {
		// instantiate new user
		var v interface{}
		// scan row result into new user instance
		err = rows.Scan(&v)
		if err != nil {
			log.Fatalf("results: %s", err)
		}
		// append res instance to results supplied set
		res = append(res, v)
	}
}

func (ds *SQLiteDataSource) SelectOneByID(table string, res interface{}, id int) {
	// sql statement
	findUser := `SELECT * FROM %s WHERE id=? LIMIT 1;`
	// execute select query
	row := ds.db.QueryRow(fmt.Sprintf(findUser, table), id)
	// scan row result into user ptr instance provided
	err := row.Scan(&res)
	if err != nil {
		log.Fatalf("results: %s", err)
	}
}

func (ds *SQLiteDataSource) Insert(table string, data []interface{}) {
	// sql statement
	insertUser := `INSERT INTO %s (first_name, last_name, email) VALUES (?,?,?);`
	// prepare to protect against sql-injection
	stmt, err := ds.db.Prepare(fmt.Sprintf(insertUser, table))
	if err != nil {
		log.Fatalf("preparing(%q): %s", insertUser, err)
	}
	// execute with params
	_, err = stmt.Exec(data...)
	if err != nil {
		log.Fatalf("executing prepared statment: %s", err)
	}
}

func (ds *SQLiteDataSource) Update(table string, data []interface{}) {
	// sql statement
	updateUser := `UPDATE %s SET first_name=?, last_name=?, email=? WHERE id=?;`
	// prepare to protect against sql-injection
	stmt, err := ds.db.Prepare(fmt.Sprintf(updateUser, table))
	if err != nil {
		log.Fatalf("preparing(%q): %s", updateUser, err)
	}
	// execute with params
	_, err = stmt.Exec(data...)
	if err != nil {
		log.Fatalf("executing prepared statment: %s", err)
	}
}

func (ds *SQLiteDataSource) Delete(table string, field string, value interface{}) {
	// sql statement
	deleteUser := `DELETE FROM %s WHERE '%s'=?;`
	// prepare to protect against sql-injection
	stmt, err := ds.db.Prepare(fmt.Sprintf(deleteUser, table, field))
	if err != nil {
		log.Fatalf("preparing(%q): %s", deleteUser, err)
	}
	// execute with params
	_, err = stmt.Exec(value)
	if err != nil {
		log.Fatalf("executing prepared statment: %s", err)
	}
}
