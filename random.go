package tbeer

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
	"time"
)

func randPos() int {
	var r int32
	binary.Read(rand.Reader, binary.LittleEndian, &r)
	if r < 0 {
		r = -r
	}
	return int(r)
}

func randFrac() float64 {
	i := randPos()
	return float64(i) / float64(0x7fffffff)
}

func errCode(e error) int {
	switch err := e.(type) {
	case *sqlite3.Error:
		return err.Code()
	default:
		return sqlite3.OK
	}
}

func randName() string {
	gr := [][]string{
		[]string{"b", "c", "d", "g", "j", "k", "p", "q", "t"}, // class 0
		[]string{"f", "h", "th", "v"},                         // class 1
		[]string{"s"},                                         // class 2
		[]string{"l", "n", "r", "w", "m"},                     // class 3
		[]string{"a", "e", "i", "o", "u", "y"},                // class 4
	}
	// state transitions: repeat for higher chance
	st := [][]int{
		[]int{3, 3, 4},
		[]int{3, 3, 4},
		[]int{1, 4, 4},
		[]int{3, 4, 4, 4},
		[]int{4, 3, 3, 2, 2, 1, 1, 1, 0, 0, 0}}
	str := ""
	state := randPos() % 5
	for i := 0; ; i++ {
		chr := gr[state][randPos()%len(gr[state])]
		str = str + chr
		if 4+(randPos()%10) < i {
			break
		}
		state = st[state][randPos()%len(st[state])]
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

type randTable struct {
	stmt *sql.Stmt
	ids  []int64
}

// Put data into table and store the id
func (t *randTable) put(args ...interface{}) {
	res, err := t.stmt.Exec(args...)
	if err != nil {
		log.Fatal(err)
	} else {
		id, _ := res.LastInsertId()
		t.ids = append(t.ids, id)
	}
}

// Put data into table, tolerate errors. Does not store id
func (t *randTable) tryPut(args ...interface{}) error {
	_, err := t.stmt.Exec(args...)
	return err
}

func (t *randTable) randId() int64 {
	return t.ids[randPos()%len(t.ids)]
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

func randPeriod(base time.Time) (time.Time, time.Time) {
	t1 := base.Add(time.Hour * time.Duration(randPos()%8760))
	t2 := t1.Add(time.Hour * time.Duration(randPos()%720))
	return t1, t2
}

func PopulateRandom() {
	rc := newRandContext()

	fmt.Println("populating random data")

	rc.db.Begin()

	user := rc.table(iq("user", []string{"alias", "login", "email"}))
	place := rc.table(iq("place", []string{"name", "lat", "long", "radius"}))
	part := rc.table(iq("participant", []string{"ownerid", "alias", "description"}))
	period := rc.table(iq("period", []string{"start", "end"}))
	meeting := rc.table(iq("meeting", []string{"ownerid", "periodid", "placeid", "name"}))
	avail := rc.table(iq("availability", []string{"ownerid", "partid", "placeid", "periodid", "description"}))
	addr := rc.table(iq("address", []string{"type", "value"}))

	mpart := rc.table(iq("meeting_participant", []string{"meetingid", "participantid"}))
	pladdr := rc.table(iq("place_address", []string{"placeid", "addressid"}))

	for i := 0; i < 20; i++ {
		user.put(randName(), "login", "test@mail.com")
	}

	for i := 0; i < 50; i++ {
		baseLat := 59.95
		baseLong := 10.75
		place.put(randName(), baseLat+(randFrac()-0.5)/10.0, baseLong+((randFrac()-0.5)/5.0), randPos()%10)
	}

	for i := 0; i < 20; i++ {
		part.put(user.randId(), randName(), "my description")
	}

	now := time.Now().Round(time.Hour)
	for i := 0; i < 20; i++ {
		t1, t2 := randPeriod(now)
		period.put(t1.Unix(), t2.Unix())
	}

	for i := 0; i < 20; i++ {
		meeting.put(user.randId(), period.randId(), place.randId(), "my meeting name")
	}

	for i := 0; i < 100; i++ {
		avail.put(user.randId(), part.randId(), place.randId(), period.randId(), "my availability reason")
	}

	for i := 0; i < 20; i++ {
		addr.put(randPos()%5, randName())
	}

	for i := 0; i < 100; i++ {
		err := mpart.tryPut(meeting.randId(), part.randId())
		switch errCode(err) {
		case 2067:
			i--
		case 0:
			continue
		default:
			log.Fatal(err)
		}
	}

	for i := 0; i < 100; i++ {
		err := pladdr.tryPut(place.randId(), addr.randId())
		switch errCode(err) {
		case 2067:
			i--
		case 0:
			continue
		default:
			log.Fatal(err)
		}
	}

	fmt.Println("committing...")
	err := rc.tx.Commit()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("... done")
	}
}
