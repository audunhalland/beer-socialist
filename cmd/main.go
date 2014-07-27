package main

import (
	"encoding/json"
	//"fmt"
	"github.com/audunhalland/beer-socialist"
	htmltmpl "html/template"
	"net/http"
	"strconv"
	texttmpl "text/template"
)

type Page struct {
	Title string
}

type TestListItem struct {
	Name string
}

/* example how to write json data */
func writeList(w http.ResponseWriter) {
	len := 5
	items := make([]TestListItem, len, len)
	for i := range items {
		items[i].Name = strconv.Itoa(i)
	}
	json.NewEncoder(w).Encode(items)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	default:
		page := &Page{Title: "index"}
		t, _ := htmltmpl.ParseFiles("./content/index.html")
		t.Execute(w, page)
	}
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	var data []byte
	var err error
	switch r.URL.Path {
	case "/json/homecoord":
		data, err = json.Marshal([2]float64{59.95, 10.75})
	}
	if err != nil {
		/* some error */
	} else {
		w.Write(data)
	}
}

func installTemplateHandler(prefix string, content_type string) {
	http.HandleFunc(prefix,
		func(w http.ResponseWriter, r *http.Request) {
			filename := "./content/" + r.URL.Path[len(prefix):]
			t, err := texttmpl.ParseFiles(filename)

			if err != nil {
				http.NotFound(w, r)
			} else {
				w.Header().Set("Content-Type", content_type)
				t.Execute(w, tbeer.GlobalEnv)
			}
		})
}

func main() {
	tbeer.LoadEnv()
	tbeer.InitDb()

	installTemplateHandler("/script/", "application/javascript")
	installTemplateHandler("/style/", "text/css")
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/json/", jsonHandler)
	http.ListenAndServe(":8080", nil)
}
