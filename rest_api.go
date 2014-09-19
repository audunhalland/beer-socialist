package tbeer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

const queueBufferSize = 0

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
type StmtRestFunc func(*DispatchContext, []*sql.Stmt, http.ResponseWriter) error

// A REST handler with prepared statements and a handler function
type StmtRestHandler struct {
	LeafDispatcher
	fn    StmtRestFunc
	stmts []*sql.Stmt
}

func (h *StmtRestHandler) ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request) {
	err := h.fn(ctx, h.stmts, w)
	if err != nil {
		jsonError(w, err)
	}
}

func installStmtRestHandler(pathPattern string, queryStrings []string, fn StmtRestFunc) {
	handler := new(StmtRestHandler)
	var err error
	handler.fn = fn
	handler.stmts, err = compileStatements(queryStrings)

	if err == nil {
		InstallRestHandler(pathPattern, handler)
	}
}

func InitRestTree() {
	installStmtRestHandler("userpref",
		[]string{
			"SELECT key, value FROM user_preference WHERE ownerid = ?",
			"SELECT value FROM user_preference WHERE ownerid = ? AND key = ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			items, err := Uniplex(queueBufferSize,
				func(out chan<- interface{}) error {
					if len(ctx.request.Form["q"]) == 0 {
						rows, err := stmts[0].Query(ctx.userid)
						if err != nil {
							return err
						}
						for rows.Next() {
							var key string
							var val interface{}
							if err := rows.Scan(&key, &val); err == nil {
								out <- &KeyedItem{key, val}
							}
						}
					} else {
						for _, p := range ctx.request.Form["q"] {
							row := stmts[1].QueryRow(ctx.userid, p)
							var val interface{}
							if err := row.Scan(&val); err == nil {
								out <- &KeyedItem{p, val}
							}
						}
					}
					return nil
				})

			if err != nil {
				return err
			}

			WriteChannelAsJSONDictionary(w, items)
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
			place := &Place{Type: "place"}
			if err := row.Scan(place.BasicFields()...); err != nil {
				return err
			} else {
				addrrows, err := stmts[1].Query(place.Id)
				place.Address = make([]*Address, 0, 10)

				if err != nil {
					fmt.Println(err)
				} else {
					for addrrows.Next() {
						addr := &Address{}
						addrrows.Scan(addr.BasicFields()...)
						place.Address = append(place.Address, addr)
					}
				}

				json.NewEncoder(w).Encode(place)
				return nil
			}
		})

	installStmtRestHandler("places",
		[]string{
			"SELECT id, name, lat, long, radius FROM place WHERE " +
				"lat > ? AND lat < ? AND long > ? AND long < ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			items, err := Uniplex(queueBufferSize,
				func(out chan<- interface{}) error {
					rect, err := GetRectangle(ctx)
					if err != nil {
						return err
					}
					rows, err := stmts[0].Query(rect.MinLat, rect.MaxLat, rect.MinLong, rect.MaxLong)
					if err != nil {
						return err
					}
					for rows.Next() {
						place := &Place{Type: "place"}
						err := rows.Scan(place.BasicFields()...)
						if err != nil {
							return err
						}
						out <- place
					}
					return nil
				})

			if err != nil {
				return err
			}

			WriteChannelAsJSONList(w, items)
			return nil
		})

	installStmtRestHandler("stuff_at",
		[]string{
			"SELECT id, name, lat, long, radius FROM place WHERE " +
				"lat > ? AND lat < ? AND long > ? AND long < ?",
			"SELECT availability.id, availability.description, " +
				"participant.id, participant.alias, participant.description, " +
				"place.id, place.name, place.lat, place.long, place.radius, " +
				"period.start, period.end " +
				"FROM availability, participant, place, period " +
				"WHERE " +
				"availability.partid = participant.id AND " +
				"availability.placeid = place.id AND " +
				"place.lat > ? AND place.lat < ? AND place.long > ? and place.long < ? AND " +
				"availability.periodid = period.id"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			rect, err := GetRectangle(ctx)
			if err != nil {
				return err
			}

			items, err := Multiplex(queueBufferSize,
				func(out chan<- interface{}) error {
					rows, err := stmts[0].Query(rect.MinLat, rect.MaxLat, rect.MinLong, rect.MaxLong)
					if err != nil {
						return err
					}
					for rows.Next() {
						place := &Place{Type: "place"}
						if err := rows.Scan(place.BasicFields()...); err != nil {
							out <- err
						} else {
							out <- place
						}
					}
					return nil
				},
				func(out chan<- interface{}) error {
					rows, err := stmts[1].Query(rect.MinLat, rect.MaxLat, rect.MinLong, rect.MaxLong)
					if err != nil {
						return err
					}
					for rows.Next() {
						a := &Availability{Type: "availability"}
						if err := rows.Scan(ConcatBasicFields(a, &a.Participant, &a.Place, &a.Period)...); err != nil {
							out <- err
						} else {
							out <- a
						}
					}
					return nil
				})

			if err != nil {
				return err
			}

			WriteChannelAsJSONList(w, items)
			return nil
		})

	installStmtRestHandler("meeting/:id",
		[]string{"SELECT id, ownerid, name FROM meeting WHERE id = ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			row := stmts[0].QueryRow(ctx.param[0])
			meeting := &Meeting{Type: "meeting"}
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
				"participant.id, participant.alias, participant.description, " +
				"place.id, place.name, place.lat, place.long, place.radius, " +
				"period.start, period.end " +
				"FROM availability, participant, place, period " +
				"WHERE " +
				"availability.ownerid = ? AND " +
				"availability.partid = participant.id AND " +
				"availability.placeid = place.id AND " +
				"availability.periodid = period.id"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			items, err := Uniplex(queueBufferSize,
				func(out chan<- interface{}) error {
					rows, err := stmts[0].Query(ctx.userid)
					if err != nil {
						return err
					}
					for rows.Next() {
						a := &Availability{Type: "availability"}
						err := rows.Scan(ConcatBasicFields(a, &a.Participant, &a.Place, &a.Period)...)
						if err != nil {
							out <- err
						} else {
							out <- a
						}
					}
					return nil
				})

			if err != nil {
				return err
			}

			WriteChannelAsJSONList(w, items)
			return nil
		})

	installStmtRestHandler("meetings",
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
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			items, err := Uniplex(queueBufferSize,
				func(out chan<- interface{}) error {
					rows, err := stmts[0].Query(ctx.userid)
					if err != nil {
						return err
					}
					for rows.Next() {
						m := &Meeting{Type: "meeting"}
						var partid int
						err := rows.Scan(append(ConcatBasicFields(m, &m.Place, &m.Period), &partid)...)
						if err != nil {
							out <- err
						} else {
							out <- m
						}
					}
					return nil
				})
			if err != nil {
				return err
			}

			WriteChannelAsJSONList(w, items)
			return nil
		})

	installStmtRestHandler("placesearch",
		[]string{"SELECT name, id FROM place WHERE name LIKE ?"},
		func(ctx *DispatchContext, stmts []*sql.Stmt, w http.ResponseWriter) error {
			type Suggestion struct {
				Value string `json:"value"`
				Data  int64  `json:"data"`
			}

			items, err := Uniplex(queueBufferSize,
				func(out chan<- interface{}) error {
					q, ok := ctx.request.Form["query"]
					if !ok {
						return fmt.Errorf("no query")
					}

					rows, err := stmts[0].Query("%" + q[0] + "%")

					if err != nil {
						return err
					}

					for rows.Next() {
						s := &Suggestion{}
						err := rows.Scan(&s.Value, &s.Data)
						if err != nil {
							out <- err
						} else {
							out <- s
						}
					}
					return nil
				})

			if err != nil {
				return err
			}

			w.Write([]byte(`{"suggestions":`))
			WriteChannelAsJSONList(w, items)
			w.Write([]byte("}"))
			return nil
		})
}
