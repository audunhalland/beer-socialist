package tbeer

import (
	"encoding/json"
	htmltmpl "html/template"
	"log"
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
				t.Execute(w, GlobalEnv)
			}
		})
}

// Start http server and block
func StartHttp() {
	installTemplateHandler("/script/", "application/javascript")
	installTemplateHandler("/style/", "text/css")

	InitRestTree()
	http.HandleFunc("/api/", HandleRestRequest)

	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/json/", jsonHandler)

	addr := ":" + strconv.FormatInt(int64(GlobalEnv.ServerPort), 10)
	var err error

	if GlobalEnv.ServerSecure {
		log.Printf("starting secure server on %s", addr)
		err = http.ListenAndServeTLS(
			addr,
			GlobalEnv.ServerCertFile,
			GlobalEnv.ServerKeyFile,
			nil)
	} else {
		log.Printf("starting non-secure server on %s", addr)
		err = http.ListenAndServe(addr, nil)
	}

	if err != nil {
		log.Fatal(err)
	}
}
