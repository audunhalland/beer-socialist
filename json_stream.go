package tbeer

import (
	"encoding/json"
	"fmt"
	"io"
)

type KeyedItem struct {
	key   string
	value interface{}
}

type EmptyList struct{}
type EmptyDictionary struct{}

func encodeItem(w io.Writer, o interface{}) error {
	bytes, err := json.Marshal(o)
	if err != nil {
		return err
	} else {
		w.Write(bytes)
		return nil
	}
}

func encodeDictionaryItem(w io.Writer, o interface{}) error {
	switch ot := o.(type) {
	case *KeyedItem:
		encodeItem(w, ot.key)
		w.Write([]byte(":"))
		encodeItem(w, ot.value)
		return nil
	case KeyedItem:
		encodeItem(w, ot.key)
		w.Write([]byte(":"))
		encodeItem(w, ot.value)
		return nil
	default:
		encodeItem(w, "error")
		w.Write([]byte(":"))
		encodeItem(w, "error")
		return fmt.Errorf("not a keyed item")
	}
}

func concatQueue(w io.Writer, queue <-chan interface{}, encode func(w io.Writer, o interface{}) error) {
	for e := range queue {
		w.Write([]byte(","))
		if encode(w, e) != nil {
			return
		}
	}
}

func encodeCollection(w io.Writer, first interface{}, rest <-chan interface{}) {
	switch ft := first.(type) {
	case EmptyList, *EmptyList:
		w.Write([]byte("[]"))
	case EmptyDictionary, *EmptyDictionary:
		w.Write([]byte("{}"))
	case *KeyedItem:
		w.Write([]byte("{"))
		encodeDictionaryItem(w, ft)
		concatQueue(w, rest, encodeDictionaryItem)
		w.Write([]byte("}"))
	default:
		w.Write([]byte("["))
		encodeItem(w, ft)
		concatQueue(w, rest, encodeItem)
		w.Write([]byte("]"))
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
				delete(firstElementMap, i)
				encodeCollection(w, first, element)
			} else {
				// No items: default to empty list
				w.Write([]byte("[]"))
			}

		default:
			encodeItem(w, element)
		}
	}

	return nil
}
