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
		"userid INTEGER, " +
		"alias TEXT, " +
		"description TEXT, " +
		"FOREIGN KEY(userid) REFERENCES user(id)" +
		")",
	"CREATE TABLE IF NOT EXISTS meeting (" +
		"id INTEGER PRIMARY KEY, " +
		"ownerid INTEGER, " +
		"name TEXT, " +
		"FOREIGN KEY(ownerid) REFERENCES user(id)" +
		")",
	"CREATE TABLE IF NOT EXISTS place (" +
		"id INTEGER PRIMARY KEY, " +
		"name TEXT, " +
		"lat REAL, " +
		"long REAL" +
		")",
	"CREATE TABLE IF NOT EXISTS meeting_participant (" +
		"id INTEGER PRIMARY KEY, " +
		"meetingid INTEGER, " +
		"participantid INTEGER, " +
		"FOREIGN KEY(meetingid) REFERENCES meeting(id), " +
		"FOREIGN KEY(participantid) REFERENCES participant(id)" +
		")",
	// Can there be many places for one meeting?
	"CREATE TABLE IF NOT EXISTS meeting_place (" +
		"id INTEGER PRIMARY KEY, " +
		"meetingid INTEGER, " +
		"placeid INTEGER, " +
		"FOREIGN KEY(meetingid) REFERENCES meeting(id), " +
		"FOREIGN KEY(placeid) REFERENCES place(id)" +
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

var GlobalDB *sql.DB

func init_table(db *sql.DB, q string) {
	res, err := db.Exec(q)
	if err != nil {
		fmt.Println(q, res, err)
	}
}

func OpenDB() (*sql.DB, error) {
	return sql.Open("sqlite3", "./tbeer.sqlite3")
}

func InitDB() {
	db, err := OpenDB()
	if err != nil {
		fmt.Println(err)
	}
	for i := range init_queries {
		init_table(db, init_queries[i])
	}
	GlobalDB = db
}

func IsDBEmpty() bool {
	rows, _ := GlobalDB.Query("SELECT count(*) FROM user")
	var count int
	rows.Next()
	rows.Scan(&count)
	rows.Close()
	return count == 0
}
