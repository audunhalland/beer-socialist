package tbeer

import (
	"bytes"
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
		{"place/1", "dict"},
		{"places", "error"}, /* missing bounding box */
		{"places?minlat=0&minlong=0&maxlat=0&maxlong=0", "list"},
		{"meeting/1", "dict"},
		{"availability", "list"},
		{"meetings", "list"},
		{"placesearch", "error"},        /* missing query */
		{"placesearch?query=a", "dict"}} /* missing query */

	OpenTestEnv()
	defer CloseTestEnv()
	serv := httptest.NewServer(RestTestHttpHandler{})
	defer serv.Close()

	for _, item := range l {
		url := serv.URL + "/api/" + item.path
		res, err := http.Get(url)
		if err != nil {
			t.Error(err)
		} else if res.StatusCode != 200 && item.expect != "error" {
			buf := &bytes.Buffer{}
			buf.ReadFrom(res.Body)
			t.Errorf("got unexpected status %d for request %s. Response: %s", res.StatusCode, url, buf.String())
		} else {
			buf := &bytes.Buffer{}
			buf.ReadFrom(res.Body)
			t.Logf("rest call OK: %s", url)
		}
	}
}

func TestSomethingElse(t *testing.T) {
	OpenTestEnv()
	defer CloseTestEnv()
}
