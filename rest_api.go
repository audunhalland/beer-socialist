package tbeer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	//"errors"
	"net/http"
	//"strconv"
)

func jsonError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(err.Error())
}

func compileStatements(q []string) ([]*sql.Stmt, error) {
	var err error
	stmts := make([]*sql.Stmt, len(q))
	for i, _ := range stmts {
		stmts[i], err = GlobalDB.Prepare(q[i])
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}
	return stmts, nil
}

// Type of the function use to handle REST requests based on one sql statement
type stmtRestFunc func(*DispatchContext, []*sql.Stmt, http.ResponseWriter) error

// A REST handler with prepared statements and a handler function
type stmtRestHandler struct {
	LeafDispatcher
	fn    stmtRestFunc
	stmts []*sql.Stmt
}

func (h *stmtRestHandler) ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request) {
	err := h.fn(ctx, h.stmts, w)
	if err != nil {
		jsonError(w, err)
	}
}

func installStmtRestHandler(pathPattern string, queryStrings []string, fn stmtRestFunc) {
	handler := new(stmtRestHandler)
	var err error
	handler.fn = fn
	handler.stmts, err = compileStatements(queryStrings)

	if err == nil {
		InstallRestHandler(pathPattern, handler)
	}
}

// Type of the function use to handle REST requests based on one sql statement
type producer func(*DispatchContext, []*sql.Stmt, chan<- interface{}) error

// A REST handler with prepared statements and a producer function
type queueRestHandler struct {
	LeafDispatcher
	prod  producer
	stmts []*sql.Stmt
}

func (h *queueRestHandler) ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request) {
	queue := make(chan interface{})

	go func() {
		err := h.prod(ctx, h.stmts, queue)
		if err != nil {
			queue <- err
		}
		close(queue)
	}()

	writeJSONList(w, queue)
}

func installQueueHandler(pathPattern string, queryStrings []string, fn producer) {
	handler := new(queueRestHandler)
	var err error
	handler.prod = fn
	handler.stmts, err = compileStatements(queryStrings)

	if err == nil {
		InstallRestHandler(pathPattern, handler)
	}
}

func writeJSONList(w http.ResponseWriter, queue <-chan interface{}) {
	enc := json.NewEncoder(w)
	e, ok := <-queue

	if ok {
		switch item := e.(type) {
		case error:
			enc.Encode("error: " + item.Error())
		default:
			w.Write([]byte("["))
			enc.Encode(item)

			for e = range queue {
				w.Write([]byte(","))
				enc.Encode(e)
			}

			w.Write([]byte("]"))
		}
	} else {
		// empty list?
		w.Write([]byte("[]"))
	}
}

func InitRestTree() {
	installStmtRestHandler("place/:id",
		[]string{
			"SELECT id, name, lat, long, radius FROM place WHERE id = ?",
			"SELECT address.type, address.value FROM address, place_address " +
				"WHERE " +
				"place_address.placeid = ? AND " +
				"place_address.addressid = address.id "},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			row := stmts[0].QueryRow(ctx.param[0])
			place := new(Place)
			var placeid int
			if err := row.Scan(append([]interface{}{&placeid}, place.BasicFields()...)...); err != nil {
				return err
			} else {
				addrrows, err := stmts[1].Query(placeid)
				place.Address = make([]*Address, 0, 10)

				if err != nil {
					fmt.Println(err)
				} else {
					for addrrows.Next() {
						addr := new(Address)
						addrrows.Scan(addr.BasicFields()...)
						place.Address = append(place.Address, addr)
					}
				}

				json.NewEncoder(w).Encode(place)
				return nil
			}
		})

	installQueueHandler("places",
		[]string{
			"SELECT name, lat, long, radius FROM place WHERE " +
				"lat > ? AND lat < ? AND long > ? AND long < ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, queue chan<- interface{}) error {
			rows, err := stmts[0].Query(-90, 90, -180, 180)
			if err != nil {
				return err
			}
			for rows.Next() {
				place := new(Place)
				err := rows.Scan(place.BasicFields()...)
				if err != nil {
					queue <- err
				} else {
					queue <- place
				}
			}
			return nil
		})

	installStmtRestHandler("meeting/:id",
		[]string{"SELECT id, ownerid, name FROM meeting WHERE id = ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			row := stmts[0].QueryRow(ctx.param[0])
			meeting := new(Meeting)
			if err := row.Scan(meeting.BasicFields()...); err != nil {
				return err
			} else {
				json.NewEncoder(w).Encode(meeting)
				return nil
			}
			return nil
		})

	installStmtRestHandler("availability",
		[]string{
			"SELECT availability.id, availability.description," +
				"participant.alias, participant.description, " +
				"place.name, place.lat, place.long, place.radius, " +
				"period.start, period.end " +
				"FROM availability, participant, place, period " +
				"WHERE " +
				"availability.ownerid = ? AND " +
				"availability.partid = participant.id AND " +
				"availability.placeid = place.id AND " +
				"availability.periodid = period.id"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			rows, err := stmts[0].Query(ctx.userid)
			if err != nil {
				return err
			}
			lst := make([]*Availability, 0)
			for rows.Next() {
				a := new(Availability)
				err := rows.Scan(ConcatBasicFields(a, &a.Participant, &a.Place, &a.Period)...)
				if err != nil {
					return err
				}
				lst = append(lst, a)
			}
			json.NewEncoder(w).Encode(lst)
			return nil
		})

	installQueueHandler("meetings",
		[]string{
			"SELECT meeting.id, meeting.ownerid, meeting.name, " +
				"place.name, place.lat, place.long, place.radius, " +
				"period.start, period.end, " +
				"participant.id " +
				"FROM meeting, place, period, meeting_participant, participant " +
				"WHERE " +
				"participant.ownerid = ? AND " +
				"meeting_participant.participantid = participant.id AND " +
				"meeting_participant.meetingid = meeting.id AND " +
				"meeting.placeid = place.id AND " +
				"meeting.periodid = period.id"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, queue chan<- interface{}) error {
			rows, err := stmts[0].Query(ctx.userid)
			if err != nil {
				return err
			}
			for rows.Next() {
				m := new(Meeting)
				var partid int
				err := rows.Scan(append(ConcatBasicFields(m, &m.Place, &m.Period), &partid)...)
				if err != nil {
					queue <- err
				} else {
					queue <- m
				}
			}
			return nil
		})

	debugRestTree(restTree, 0)
}
