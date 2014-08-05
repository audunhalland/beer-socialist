package tbeer

import (
	"encoding/json"
	"fmt"
	"io"
)

func encode(w io.Writer, o interface{}) {
	bytes, err := json.Marshal(o)
	if err != nil {
		fmt.Println(err)
	} else {
		w.Write(bytes)
	}
}

// Ability to encode json lists as streams using channels
// The concatenated list of elements make up the body of the json document.
// Supported types:
// byte[]: will be used directly
// <-chan interface{} will be written as a list
func StreamEncodeJSON(w io.Writer, elements []interface{}) error {
	firstElementMap := make(map[int]interface{})

	// initialize.
	// collect the first items of potential streams and complain if there is
	// an error.
	for i, ei := range elements {
		switch element := ei.(type) {
		case <-chan interface{}:
			e, ok := <-element

			if ok {
				switch item := e.(type) {
				case error:
					return item
				default:
					// store item
					firstElementMap[i] = item
				}
			}
		}
	}

	// stream to output
	for i, ei := range elements {
		switch element := ei.(type) {
		case []byte:
			w.Write(element)
		case <-chan interface{}:
			first, exists := firstElementMap[i]

			if exists {
				w.Write([]byte("["))
				encode(w, first)
				delete(firstElementMap, i)

				for e := range element {
					w.Write([]byte(","))
					encode(w, e)
				}

				w.Write([]byte("]"))
			} else {
				// empty
				w.Write([]byte("[]"))
			}

		default:
			encode(w, element)
		}
	}

	return nil
}
