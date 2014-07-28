package tbeer

import (
	//sqlite3 "code.google.com/p/go-sqlite/go1/sqlite3"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strings"
)

type randTable struct {
	stmt *sql.Stmt
	ids  []int64
}

func (t *randTable) put(args ...interface{}) {
	res, err := t.stmt.Exec(args...)
	if err != nil {
		fmt.Println(err)
	} else {
		id, _ := res.LastInsertId()
		t.ids = append(t.ids, id)
	}
}

func (t *randTable) randId() int64 {
	var r int
	binary.Read(rand.Reader, binary.LittleEndian, &r)
	return t.ids[r%len(t.ids)]
}

type randContext struct {
	db *sql.DB
	tx *sql.Tx
}

func (rc *randContext) table(q string) *randTable {
	fmt.Println(q)
	stmt, err := rc.db.Prepare(q)
	if err != nil {
		fmt.Println(err)
		return nil
	} else {
		return &randTable{rc.tx.Stmt(stmt), make([]int64, 0, 1000)}
	}
}

func newRandContext() *randContext {
	rc := new(randContext)
	rc.db = GlobalDB
	rc.tx, _ = rc.db.Begin()
	return rc
}

// return a slice that contains n copies of str
// BUG: isn't there a more elegant way of doing this in Go?
func repeat(str string, n int) []string {
	ret := make([]string, n)
	for i := 0; i < n; i++ {
		ret[i] = str
	}
	return ret
}

// generate an INSERT INTO query
func iq(table string, columns []string) string {
	return "INSERT INTO " + table + " (" +
		strings.Join(columns, ",") + ") VALUES (" +
		strings.Join(repeat("?", len(columns)), ",") + ")"
}

func PopulateRandom() {
	rc := newRandContext()

	fmt.Println("populating random data")

	rc.db.Begin()

	user := rc.table(iq("user", []string{"alias", "login", "email"}))
	place := rc.table(iq("place", []string{"name", "lat", "long"}))
	part := rc.table(iq("participant", []string{"userid", "alias", "description"}))
	meeting := rc.table(iq("meeting", []string{"name"}))
	mpart := rc.table(iq("meeting_participant", []string{"meetingid", "participantid"}))
	mplace := rc.table(iq("meeting_place", []string{"meetingid", "placeid"}))

	for i := 0; i < 10; i++ {
		user.put("yo", "login", "test@mail.com")
	}

	for i := 0; i < 10; i++ {
		place.put("place", 59.95+(float64(i)/100.0), 10.75+(float64(i)/100.0))
	}

	for i := 0; i < 10; i++ {
		part.put(user.randId(), "participant alias", "participant description")
	}

	for i := 0; i < 10; i++ {
		meeting.put("meeting name")
	}

	for i := 0; i < 20; i++ {
		mpart.put(meeting.randId(), part.randId())
	}

	for i := 0; i < 20; i++ {
		mplace.put(meeting.randId(), place.randId())
	}

	fmt.Println("committing...")
	err := rc.tx.Commit()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("... done")
	}
}
