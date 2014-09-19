package tbeer

import (
	"encoding/json"
	"fmt"
	"io"
)

// KeyedItem is used in interface channels to signify that something
// is to be part of a json dictionary instead of a list.
type KeyedItem struct {
	key   string
	value interface{}
}

// EmptyDictionary is the only way to send an empty dictionary
// over a channel.
type EmptyDictionary struct{}

// A value of this instance indicates an empty list.
// It's here for completeness and not strictly needed,
// because a closed channel with no objects sent
// is an empty list by default
type EmptyList struct{}

// Encode any value as json
func encodeItem(w io.Writer, o interface{}) error {
	bytes, err := json.Marshal(o)
	if err != nil {
		return err
	} else {
		w.Write(bytes)
		return nil
	}
}

// Encode one dictionary item (KeyedItem)
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

// Output all collection items preceded by comma
func appendCollectionItems(w io.Writer, queue <-chan interface{}, encode func(w io.Writer, o interface{}) error) {
	for e := range queue {
		w.Write([]byte(","))
		if encode(w, e) != nil {
			return
		}
	}
}

// Encode a collection from a channel. Supports both list and dictionary.
func encodeCollection(w io.Writer, first interface{}, rest <-chan interface{}) {
	switch ft := first.(type) {
	case EmptyList, *EmptyList:
		w.Write([]byte("[]"))
	case EmptyDictionary, *EmptyDictionary:
		w.Write([]byte("{}"))
	case *KeyedItem:
		w.Write([]byte("{"))
		encodeDictionaryItem(w, ft)
		appendCollectionItems(w, rest, encodeDictionaryItem)
		w.Write([]byte("}"))
	default:
		w.Write([]byte("["))
		encodeItem(w, ft)
		appendCollectionItems(w, rest, encodeItem)
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
	// an error. An error in the first collection element will terminate
	// the encoder.
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

func WriteChannelAsJSONList(w io.Writer, items <-chan interface{}) {
	w.Write([]byte("["))
	if item, ok := <-items; ok {
		encodeItem(w, item)
		for item := range items {
			w.Write([]byte(","))
			if err := encodeItem(w, item); err != nil {
				w.Write([]byte("null"))
				break
			}
		}
	}
	w.Write([]byte("]"))
}

func WriteChannelAsJSONDictionary(w io.Writer, items <-chan interface{}) {
	w.Write([]byte("{"))
	if item, ok := <-items; ok {
		encodeDictionaryItem(w, item)
		for item := range items {
			w.Write([]byte(","))
			encodeDictionaryItem(w, item)
		}
	}
	w.Write([]byte("}"))
}
