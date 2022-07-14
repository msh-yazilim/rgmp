package scan

import (
	"encoding/json"
	"fmt"
)

type req struct {
	Serial string
	Size   int
	Msg    *msg
}

// returns in json string
func (r req) String() string {

	buff, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprint(err)
	}

	return string(buff)
}

// Message
//
// tag is 3 bytes wide
//
// fields are public to avoid reserved words
// msg itself is private anyway
type msg struct {
	Type   string
	Len    int
	Groups []*group
}

// Group
//
// tag is 2 bytes wide
type group struct {
	Type   string
	Len    int
	Fields []*field
}

// Field
//
// tag is 3 bytes wide
type field struct {
	Type string
	Len  int
	// hand over to parser
	Value []byte
}

func (f *field) MarshalJSON() ([]byte, error) {
	v := struct {
		Type  string
		Len   int
		Value string
	}{
		Type:  f.Type,
		Len:   f.Len,
		Value: printSeperateHex(f.Value),
	}

	return json.Marshal(v)
}
