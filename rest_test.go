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
		{"userpref", "dict"},
		{"userpref?q=homelat", "dict"},
		{"place/1", "dict"},
		{"places", "error"}, /* missing bounding box */
		{"places?minlat=abcde", "error"},
		{"places?minlat=-90&minlong=-180&maxlat=90&maxlong=180", "list"},
		{"meeting/1", "dict"},
		{"availability", "list"},
		{"meetings", "list"},
		{"placesearch", "error"}, /* missing query */
		{"placesearch?query=a", "dict"}}

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
