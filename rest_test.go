package tbeer

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RestTestHttpHandler struct{}

func (s RestTestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	HandleRestRequest(w, r)
}

// traverse rest tree and test that none of the calls produce errors
func TestRestGetError(t *testing.T) {
	type Expect struct {
		path   string
		expect string
	}
	l := []Expect{
		{"place/1", "yo"},
		{"places", "yo"},
		{"meeting/1", "yo"},
		{"availability", "yo"},
		{"meetings", "yo"},
		{"placesearch", "yo"}}

	OpenTestEnv()
	defer CloseTestEnv()
	serv := httptest.NewServer(RestTestHttpHandler{})
	defer serv.Close()

	for _, item := range l {
		url := serv.URL + "/api/" + item.path
		res, err := http.Get(url)
		if err != nil {
			t.Error(err)
		} else if res.StatusCode != 200 {
			t.Errorf("got status %d for request %s", res.StatusCode, url)
		} else {
			buf := &bytes.Buffer{}
			buf.ReadFrom(res.Body)
			fmt.Println(buf.String())
		}
	}
}

func TestSomethingElse(t *testing.T) {
	OpenTestEnv()
	defer CloseTestEnv()
}
