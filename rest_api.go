package tbeer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const queueBufferSize = 0

type Rectangle struct {
	MinLat  float64
	MinLong float64
	MaxLat  float64
	MaxLong float64
}

func jsonError(w http.ResponseWriter, err error) {
	w.WriteHeader(400)
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

// Function type for api calls returning a list of objects
type itemProducer func(*DispatchContext, []*sql.Stmt, chan<- interface{}) error

// A REST handler with prepared statements and a producer function
type streamRestHandler struct {
	LeafDispatcher
	elements []interface{}
	stmts    []*sql.Stmt
}

func makeItemQueue(ctx *DispatchContext, stmts []*sql.Stmt, p itemProducer) <-chan interface{} {
	queue := make(chan interface{}, queueBufferSize)

	go func() {
		err := p(ctx, stmts, queue)
		if err != nil {
			queue <- err
		}
		close(queue)
	}()

	return queue
}

func (h *streamRestHandler) ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request) {
	mappedElements := make([]interface{}, len(h.elements))

	for i, e := range h.elements {
		switch element := e.(type) {
		case func(*DispatchContext, []*sql.Stmt, chan<- interface{}) error:
			mappedElements[i] = makeItemQueue(ctx, h.stmts, element)
		default:
			mappedElements[i] = element
		}
	}

	err := StreamEncodeJSON(w, mappedElements)

	if err != nil {
		jsonError(w, err)
	}
}

// Install an api call that returns a list of objects, using a queue implemented with a channel
// Stream elements is a list of elements to encode in the json stream.
// If an element is a "producer" function, it will translate into a channel for list encoding
func installStreamHandler(pathPattern string, queryStrings []string, streamElements ...interface{}) {
	handler := new(streamRestHandler)
	var err error
	handler.elements = streamElements
	handler.stmts, err = compileStatements(queryStrings)

	if err == nil {
		InstallRestHandler(pathPattern, handler)
	}
}

func getFormFloat(m url.Values, key string) (float64, error) {
	val, ok := m[key]
	if !ok {
		return 0.0, fmt.Errorf("missing key %s", key)
	}
	f, err := strconv.ParseFloat(val[0], 64)
	if err != nil {
		return 0.0, fmt.Errorf("could not parse number: %s", val[0])
	} else {
		return f, nil
	}
}

func getRectangle(ctx *DispatchContext) (*Rectangle, error) {
	r := new(Rectangle)
	var err error
	keys := []string{"minlat", "minlong", "maxlat", "maxlong"}
	targets := []*float64{&r.MinLat, &r.MinLong, &r.MaxLat, &r.MaxLong}
	for i, t := range targets {
		*t, err = getFormFloat(ctx.request.Form, keys[i])
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func InitRestTree() {
	installStreamHandler("userpref",
		[]string{
			"SELECT key, value FROM user_preference WHERE ownerid = ?",
			"SELECT value FROM user_preference WHERE ownerid = ? AND key = ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, queue chan<- interface{}) error {
			count := 0
			if len(ctx.request.Form["q"]) == 0 {
				rows, err := stmts[0].Query(ctx.userid)
				if err != nil {
					return err
				}
				for rows.Next() {
					var key string
					var val interface{}
					if err := rows.Scan(&key, &val); err == nil {
						queue <- &KeyedItem{key, val}
						count++
					}
				}
			} else {
				for _, p := range ctx.request.Form["q"] {
					row := stmts[1].QueryRow(ctx.userid, p)
					var val interface{}
					if err := row.Scan(&val); err == nil {
						queue <- &KeyedItem{p, val}
						count++
					}
				}
			}
			if count == 0 {
				queue <- &EmptyDictionary{}
			}
			return nil
		})

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
			if err := row.Scan(place.BasicFields()...); err != nil {
				return err
			} else {
				addrrows, err := stmts[1].Query(place.Id)
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

	installStreamHandler("places",
		[]string{
			"SELECT id, name, lat, long, radius FROM place WHERE " +
				"lat > ? AND lat < ? AND long > ? AND long < ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, queue chan<- interface{}) error {
			rect, err := getRectangle(ctx)
			if err != nil {
				return err
			}
			rows, err := stmts[0].Query(rect.MinLat, rect.MaxLat, rect.MinLong, rect.MaxLong)
			if err != nil {
				return err
			}
			for rows.Next() {
				place := new(Place)
				err := rows.Scan(place.BasicFields()...)
				if err != nil {
					return err
				}
				queue <- place
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

	installStreamHandler("availability",
		[]string{
			"SELECT availability.id, availability.description," +
				"participant.alias, participant.description, " +
				"place.id, place.name, place.lat, place.long, place.radius, " +
				"period.start, period.end " +
				"FROM availability, participant, place, period " +
				"WHERE " +
				"availability.ownerid = ? AND " +
				"availability.partid = participant.id AND " +
				"availability.placeid = place.id AND " +
				"availability.periodid = period.id"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, queue chan<- interface{}) error {
			rows, err := stmts[0].Query(ctx.userid)
			if err != nil {
				return err
			}
			for rows.Next() {
				a := new(Availability)
				err := rows.Scan(ConcatBasicFields(a, &a.Participant, &a.Place, &a.Period)...)
				if err != nil {
					return err
				}
				queue <- a
			}
			return nil
		})

	installStreamHandler("meetings",
		[]string{
			"SELECT meeting.id, meeting.ownerid, meeting.name, " +
				"place.id, place.name, place.lat, place.long, place.radius, " +
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
					return err
				}
				queue <- m
			}
			return nil
		})

	installStreamHandler("placesearch",
		[]string{"SELECT name, id FROM place WHERE name LIKE ?"},
		[]byte(`{"suggestions":`),
		func(ctx *DispatchContext, stmts []*sql.Stmt, queue chan<- interface{}) error {
			type Suggestion struct {
				Value string `json:"value"`
				Data  int64  `json:"data"`
			}
			q, ok := ctx.request.Form["query"]
			if !ok {
				return fmt.Errorf("no query")
			}

			rows, err := stmts[0].Query("%" + q[0] + "%")

			if err != nil {
				return err
			}

			for rows.Next() {
				s := new(Suggestion)
				err := rows.Scan(&s.Value, &s.Data)
				if err != nil {
					return err
				}
				queue <- s
			}
			return nil
		},
		[]byte("}"))
}
