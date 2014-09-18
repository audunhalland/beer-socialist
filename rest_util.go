package tbeer

import (
	"fmt"
	"net/url"
	"strconv"
)

type Rectangle struct {
	MinLat  float64
	MinLong float64
	MaxLat  float64
	MaxLong float64
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

// Extract a rectangle from dispatched rest request
func GetRectangle(ctx *DispatchContext) (*Rectangle, error) {
	r := &Rectangle{}
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
