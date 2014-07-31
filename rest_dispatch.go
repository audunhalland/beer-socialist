package tbeer

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type DispatchContext struct {
	// userid of the dispatched request
	userid int64
	param  []interface{}
}

type dispatcher interface {
	dispatch(item string, context *DispatchContext) (dispatcher, error)
	install(key string, dp dispatcher) error
}

type RESTHandler interface {
	ServeREST(ctx *DispatchContext, w http.ResponseWriter, r *http.Request)
}

// leaf dispatcher should also implement RESTHandler
type LeafDispatcher struct {
}

// dispatcher that searches for a corresponding child dispatcher
type selectDP struct {
	children map[string]dispatcher
}

// dispatcher that accepts the level passed
type acceptDP struct {
	child dispatcher
}

// dispatcher that reads an integer
type intDP struct {
	acceptDP
}

// dispatcher that reads a string
type stringDP struct {
	acceptDP
}

func newSelectDP() *selectDP {
	return &selectDP{make(map[string]dispatcher)}
}

func (s *selectDP) dispatch(item string, ctx *DispatchContext) (dispatcher, error) {
	dp, present := s.children[item]
	if present {
		return dp, nil
	} else {
		return nil, fmt.Errorf("not found: %s", item)
	}
}

func (s *selectDP) install(key string, dp dispatcher) error {
	s.children[key] = dp
	return nil
}

// Accept the item and go to child
func (s *acceptDP) dispatch(item string, ctx *DispatchContext) (dispatcher, error) {
	fmt.Println("accept dispatch. child is", s.child)
	return s.child, nil
}

func (s *acceptDP) install(key string, dp dispatcher) error {
	if s.child != nil {
		return errors.New("accept dispatcher child already exists")
	} else {
		s.child = dp
		return nil
	}
}

func (s *LeafDispatcher) dispatch(item string, ctx *DispatchContext) (dispatcher, error) {
	return nil, nil
}

func (s *LeafDispatcher) install(key string, dp dispatcher) error {
	return errors.New("can't install in a leaf")
}

func (s *intDP) dispatch(item string, ctx *DispatchContext) (dispatcher, error) {
	i, err := strconv.ParseInt(item, 10, 64)
	if err != nil {
		return nil, err
	}
	ctx.param = append(ctx.param, i)
	return s.acceptDP.dispatch(item, ctx)
}

func (s *stringDP) dispatch(item string, ctx *DispatchContext) (dispatcher, error) {
	ctx.param = append(ctx.param, item)
	return s.acceptDP.dispatch(item, ctx)
}

var restTree = newSelectDP()

func dispatchRESTPath(path string) (RESTHandler, *DispatchContext, error) {
	ctx := new(DispatchContext)
	var dis dispatcher = restTree
	remain := path
	for dis != nil {
		spl := strings.SplitN(remain, "/", 2)
		next, err := dis.dispatch(spl[0], ctx)
		if err != nil {
			fmt.Println(err)
			return nil, nil, err
		} else if next == nil {
			// now we should have a rest handler
			handler, ok := dis.(RESTHandler)
			if ok {
				return handler, ctx, nil
			} else {
				return nil, nil, errors.New("no handler")
			}
		}
		dis = next
		if len(spl) > 1 {
			remain = spl[1]
		} else {
			remain = ""
		}
	}
	return nil, nil, errors.New("impossible")
}

func HandleRestRequest(w http.ResponseWriter, r *http.Request) {
	handler, ctx, err := dispatchRESTPath(r.URL.Path[len("/api/"):])
	if err != nil {
		http.NotFound(w, r)
	} else {
		w.Header().Set("Content-Type", "application/json")

		// BUG: faking the user id of the context, which would be supplied
		// using a token in the future
		ctx.userid = 1

		handler.ServeREST(ctx, w, r)
	}
}

func InstallRestHandler(pathPattern string, restHandler dispatcher) {
	elements := strings.Split(strings.TrimRight(pathPattern, "/"), "/")
	var parent dispatcher = restTree

	for i := 1; i < len(elements); i++ {
		var dp dispatcher
		element := elements[i]
		if strings.HasPrefix(element, ":") {
			dp = new(intDP)
		} else {
			dp = newSelectDP()
		}
		err := parent.install(elements[i-1], dp)
		if err != nil {
			panic(err)
		}
		parent = dp
	}

	err := parent.install(elements[len(elements)-1], restHandler)
	if err != nil {
		panic(err)
	}
}

func debugRestTree(dp interface{}, level int) {
	ind := func() string {
		return strings.Repeat(" ", level)
	}
	if dp == nil {
		return
	}
	switch dyn := dp.(type) {
	case *selectDP:
		fmt.Println(ind() + "selectDP")
		for key, d := range dyn.children {
			fmt.Println(ind() + key + ":")
			debugRestTree(d, level+1)
		}
	case *intDP:
		fmt.Println(ind() + "int")
		debugRestTree(dyn.child, level+1)
	case RESTHandler:
		fmt.Println(ind() + "http")
	default:
		fmt.Println(ind() + "unknown")
	}
}
