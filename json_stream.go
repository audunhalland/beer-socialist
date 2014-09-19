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
