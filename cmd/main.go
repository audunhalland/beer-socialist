package main

import (
	"encoding/json"
	//"fmt"
	"github.com/audunhalland/tbeer"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Page struct {
	Title string
}

type TestListItem struct {
	Name string
}

func writeRaw(w http.ResponseWriter, data []byte) {
    w.Write(data)
}

func writeList(w http.ResponseWriter) {
	len := 5
	items := make([]TestListItem, len, len)
	for i := range items {
		items[i].Name = strconv.Itoa(i)
	}
	data, _ := json.Marshal(items)
	w.Write(data)
}

func writeStyle(w http.ResponseWriter, data []byte) {
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	default:
		page := &Page{Title: "index"}
		t, _ := template.ParseFiles("./content/index.html")
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

func installFileHandler(prefix string, fn func(http.ResponseWriter, []byte), content_type string) {
	http.HandleFunc(prefix,
		func(w http.ResponseWriter, r *http.Request) {
            var filename = "./content/" + r.URL.Path[len(prefix):]
			var data, err = ioutil.ReadFile(filename)

			if err != nil {
				http.NotFound(w, r)
			} else {
                w.Header().Set("Content-Type", content_type)
				fn(w, data)
			}
		})
}

func main() {
	tbeer.InitDb()

	installFileHandler("/script/", writeRaw, "application/javascript")
	installFileHandler("/style/", writeRaw, "text/css")
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/json/", jsonHandler)
	http.ListenAndServe(":8080", nil)
}
