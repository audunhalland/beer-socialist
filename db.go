package tbeer

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
	"fmt"
)

var init_queries = [...]string{
	"CREATE TABLE IF NOT EXISTS user (" +
		"id INTEGER PRIMARY KEY, " +
		"alias TEXT, " +
		"login TEXT, " +
		"email TEXT " +
		")",
	"CREATE TABLE IF NOT EXISTS participant (" +
		"id INTEGER PRIMARY KEY, " +
		"user INTEGER, " +
		"alias TEXT, " +
		"description TEXT, " +
		"FOREIGN KEY(userid) REFERENCES user(id)" +
		")",
	"CREATE TABLE IF NOT EXISTS meeting (" +
		"id INTEGER PRIMARY KEY, " +
		"name TEXT, " +
		"lat REAL, " +
		"long REAL" +
		")",
	"CREATE TABLE IF NOT EXISTS place (" +
		"id INTEGER PRIMARY KEY, " +
		"name TEXT, " +
		"lat REAL, " +
		"long REAL" +
		")",
	"CREATE TABLE IF NOT EXISTS user_review (" +
		"id INTEGER PRIMARY KEY, " +
		"reviewer_id INTEGER, " +
		"reviewee_id INTEGER, " +
		"meeting_id INTEGER, " +
		"score INTEGER, " +
		"FOREIGN KEY(reviewer_id) REFERENCES user(id), " +
		"FOREIGN KEY(reviewee_id) REFERENCES user(id) " +
		")",
}

func init_table(db *sql.DB, q string) {
	res, err := db.Exec(q)
	if err != nil {
		fmt.Println(q, res, err)
	}
}

func InitDb() {
	db, err := sql.Open("sqlite3", "./tbeer.sqlite3")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("db: %p", db)
	for i := range init_queries {
		init_table(db, init_queries[i])
	}
	fmt.Println("db init done")
	db.Close()
}
