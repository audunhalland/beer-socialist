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

// Type of the function use to handle REST requests based on one sql statement
type stmtRestFunc func(*DispatchContext, *sql.Stmt, http.ResponseWriter) error

// A REST handler with one prepared statement and a handler function
type stmtRestHandler struct {
	LeafDispatcher
	fn   stmtRestFunc
	stmt *sql.Stmt
}

func (h *stmtRestHandler) ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request) {
	err := h.fn(ctx, h.stmt, w)
	if err != nil {
		jsonError(w, err)
	}
}

func installStmtRestHandler(pathPattern string, queryString string, fn stmtRestFunc) {
	handler := new(stmtRestHandler)
	var err error
	handler.fn = fn
	handler.stmt, err = GlobalDB.Prepare(queryString)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("installing ", pathPattern)
		InstallRestHandler(pathPattern, handler)
	}
}

// Handler for /place/:id/
func placeHandler(ctx *DispatchContext, stmt *sql.Stmt, w http.ResponseWriter) error {
	row := stmt.QueryRow(ctx.param[0])
	place := new(Place)
	if err := place.LoadBasic(row); err != nil {
		return err
	} else {
		json.NewEncoder(w).Encode(place)
		return nil
	}
}

// Handler for /meeting/:id/
func meetingHandler(ctx *DispatchContext, stmt *sql.Stmt, w http.ResponseWriter) error {
	row := stmt.QueryRow(ctx.param[0])
	meeting := new(Meeting)
	if err := meeting.LoadBasic(row); err != nil {
		return err
	} else {
		json.NewEncoder(w).Encode(meeting)
		return nil
	}
}

func InitRestTree() {
	installStmtRestHandler("place/:id/",
		"SELECT name, lat, long FROM place WHERE id = ?", placeHandler)
	installStmtRestHandler("meeting/:id/",
		"SELECT ownerid, name FROM meeting WHERE id = ?", meetingHandler)

	debugRestTree(restTree, 0)
}
