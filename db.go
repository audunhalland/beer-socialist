package tbeer

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
	"fmt"
)

// A BasicFieldContainer is something that contains
// fields that are basic (i.e. do not point to other containers)
type BasicFieldContainer interface {
	// Get all fields that do not point to other containers
	BasicFields() []interface{}
}

// Get the full concatenated list of BasicFields from the input list
// of field containers
func ConcatBasicFields(lst ...BasicFieldContainer) []interface{} {
	f := make([]interface{}, 0)
	for _, c := range lst {
		f = append(f, c.BasicFields()...)
	}
	return f
}

// In memory representation: Place
type Place struct {
	Type    string /* BUG: for json */
	Id      int64
	Name    string
	Lat     float64
	Long    float64
	Radius  int
	Address []*Address
}

func (s *Place) BasicFields() []interface{} {
	return []interface{}{&s.Id, &s.Name, &s.Lat, &s.Long, &s.Radius}
}

type MeetingParticipant struct {
}

// In memory representation: Meeting
type Meeting struct {
	Type         string /* BUG: for json */
	Id           int64
	Owner        int64
	Name         string
	Place        Place
	Period       Period
	Participants []MeetingParticipant
}

func (m *Meeting) BasicFields() []interface{} {
	return []interface{}{&m.Id, &m.Owner, &m.Name}
}

// In memory representation: Availability
type Availability struct {
	Type        string /* BUG: for json */
	Id          int64
	Description string
	Participant Participant
	Place       Place
	Period      Period
}

func (a *Availability) BasicFields() []interface{} {
	return []interface{}{&a.Id, &a.Description}
}

// In memory representation: Participant
type Participant struct {
	Id          int64
	Alias       string
	Description string
}

func (p *Participant) BasicFields() []interface{} {
	return []interface{}{&p.Id, &p.Alias, &p.Description}
}

// In memory representation: Period
type Period struct {
	Start int
	End   int
}

func (p *Period) BasicFields() []interface{} {
	return []interface{}{&p.Start, &p.End}
}

type Address struct {
	Type  int
	Value string
}

func (a *Address) BasicFields() []interface{} {
	return []interface{}{&a.Type, &a.Value}
}

var init_queries = [...]string{
	"CREATE TABLE IF NOT EXISTS user (" +
		"id INTEGER PRIMARY KEY, " +
		"alias TEXT, " +
		"login TEXT, " +
		"email TEXT " +
		")",
	"CREATE TABLE IF NOT EXISTS user_preference (" +
		"ownerid INTEGER NOT NULL, " +
		"key TEXT NOT NULL, " +
		"value, " +
		"FOREIGN KEY(ownerid) REFERENCES user(id)" +
		"PRIMARY KEY(ownerid, key)" +
		")",
	"CREATE TABLE IF NOT EXISTS participant (" +
		"id INTEGER PRIMARY KEY, " +
		"ownerid INTEGER NOT NULL, " +
		"alias TEXT, " +
		"description TEXT, " +
		"FOREIGN KEY(ownerid) REFERENCES user(id)" +
		")",
	"CREATE TABLE IF NOT EXISTS period (" +
		"id INTEGER PRIMARY KEY, " +
		"start INTEGER, " +
		"end INTEGER" +
		")",
	"CREATE TABLE IF NOT EXISTS meeting (" +
		"id INTEGER PRIMARY KEY, " +
		"ownerid INTEGER NOT NULL, " +
		"periodid INTEGER NOT NULL, " +
		"placeid INTEGER NOT NULL, " +
		"name TEXT, " +
		"FOREIGN KEY(ownerid) REFERENCES user(id), " +
		"FOREIGN KEY(periodid) REFERENCES period(id)" +
		")",
	"CREATE TABLE IF NOT EXISTS place (" +
		"id INTEGER PRIMARY KEY, " +
		"name TEXT, " +
		"lat REAL, " +
		"long REAL, " +
		// big or small? e.g.
		// (continent > country > county > city > neighbourhood > ... > "addressable")
		"radius INTEGER, " +
		"timezone TEXT" +
		")",
	"CREATE TABLE IF NOT EXISTS meeting_participant (" +
		"meetingid INTEGER NOT NULL, " +
		"participantid INTEGER NOT NULL, " +
		"FOREIGN KEY(meetingid) REFERENCES meeting(id), " +
		"FOREIGN KEY(participantid) REFERENCES participant(id), " +
		"PRIMARY KEY(meetingid, participantid)" +
		//") WITHOUT ROWID", requires sqlite version 3.8.2
		")",
	// availability - a period in which a meeting participant is available
	"CREATE TABLE IF NOT EXISTS availability (" +
		"id INTEGER PRIMARY KEY, " +
		"ownerid INTEGER NOT NULL, " +
		"partid INTEGER NOT NULL, " +
		"placeid INTEGER NOT NULL, " +
		"periodid INTEGER NOT NULL, " +
		"description TEXT, " +
		"FOREIGN KEY(ownerid) REFERENCES meeting(user), " +
		"FOREIGN KEY(partid) REFERENCES meeting_participant(id), " +
		"FOREIGN KEY(placeid) REFERENCES place(id), " +
		"FOREIGN KEY(periodid) REFERENCES period(id)" +
		")",
	"CREATE TABLE IF NOT EXISTS address (" +
		"id INTEGER PRIMARY KEY, " +
		"type INTEGER NOT NULL, " +
		"value TEXT" +
		")",
	"CREATE TABLE IF NOT EXISTS place_address (" +
		"placeid INTEGER NOT NULL, " +
		"addressid INTEGER NOT NULL, " +
		"FOREIGN KEY(placeid) REFERENCES place(id), " +
		"FOREIGN KEY(addressid) REFERENCES address(id), " +
		"PRIMARY KEY(placeid, addressid) " +
		//") WITHOUT ROWID", requires sqlite version 3.8.2
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
	"CREATE TABLE IF NOT EXISTS dynamic_url (" +
		"value TEXT PRIMARY KEY NOT NULL, " +
		"type INTEGER NOT NULL, " +
		"foreignid INTEGER NOT NULL " +
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
