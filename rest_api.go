package tbeer

import (
	//	"database/sql"
	"encoding/json"
	//"fmt"
	"errors"
	"net/http"
	"strconv"
)

func jsonError(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(err.Error())
}

type defaultRestHandler struct {
	LeafDispatcher
}

func (*defaultRestHandler) ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("yadayada " + strconv.FormatInt(ctx.param[0].(int64), 10)))
}

type placeRestHandler struct {
	LeafDispatcher
}

type dbPlace struct {
	Name string
	Lat  float64
	Long float64
}

func (*placeRestHandler) ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request) {
	// BUG: use precompiled statements, and a generic way of adding new rest requests
	rows, err := GlobalDB.Query("SELECT name, lat, long FROM place WHERE id = ?", ctx.param[0])

	if err != nil {
		jsonError(w, err)
	}

	if rows.Next() {
		dbPlace := new(dbPlace)
		err = rows.Scan(&dbPlace.Name, &dbPlace.Lat, &dbPlace.Long)
		json.NewEncoder(w).Encode(&dbPlace)
		rows.Next()
	} else {
		jsonError(w, errors.New("not found"))
	}
}

func InitRestTree() {
	InstallRestHandler("places/:id/", new(placeRestHandler))
	InstallRestHandler("meetings/:id/", new(defaultRestHandler))
	InstallRestHandler("users/:id/", new(defaultRestHandler))

	debugRestTree(restTree, 0)
}
