package scan

import "fmt"

// req. scanner
//
// use NewScanner for an instance
type Scanner struct {
	src  []byte
	curr int
	req  *req // message to be returned
}

const (
	MessageTagLen = 3
	GroupTagLen   = 2
	FieldTagLen   = 3
)

const MaxMsgSize = 2048

func NewScanner(src []byte) *Scanner {
	s := new(Scanner)
	s.src = src
	s.curr = 0

	return s
}

// Scans request
//
// req: size(2 bytes), tpdu(5 bytes),serial(12 bytes), msg , LRC
func (s *Scanner) Scan() (*req, error) {
	s.req = new(req)

	// first two bytes are size,
	// in big endian
	lbyte := s.step()
	left := int(lbyte)
	left = left << 8
	right := int(s.step())
	size := left + right

	// msg = (size and lrc bytes are discarded)
	msg := s.src[s.curr : len(s.src)-1]

	// size
	if size != len(msg) {
		return nil, fmt.Errorf("parsed size is not equal request's length")
	}

	if size > MaxMsgSize {
		return nil, fmt.Errorf("message size exceeds the max.")
	}

	// LRC
	lrc := s.src[len(s.src)-1]

	sum := byte(0)
	for _, b := range msg {
		sum ^= b
	}
	if sum != lrc {
		return nil, fmt.Errorf("corrupted data, lrc failed")
	}

	// tpdu
	if tpdu := s.step(); tpdu != 0x60 {
		return nil, fmt.Errorf("hatali tpdu, parse error: %x", tpdu)
	}
	// skip remaining
	s.advance(4)

	s.req.Serial = string(s.advance(12))
	s.scanMsg()
	s.req.Size = size
	return s.req, nil
}

// msg: type , len,  group*
func (s *Scanner) scanMsg() {
	s.req.Msg = new(msg)

	msgType := s.advance(MessageTagLen)
	s.req.Msg.Type = btoa(msgType)
	s.req.Msg.Len = s.scanLen()

	s.scanGroups()
}

// groups: group*
// group: type,len, field*
func (s *Scanner) scanGroups() {
	groups := make([]*group, 0)
	end := s.curr + s.req.Msg.Len

	for s.curr < end {
		g := new(group)
		g.Type = btoa(s.advance(GroupTagLen))
		g.Len = s.scanLen()
		s.scanFields(g)

		groups = append(groups, g)
	}

	s.req.Msg.Groups = groups
}

// fields: fields*
// field: type, len
func (s *Scanner) scanFields(g *group) {
	fields := make([]*field, 0)
	end := s.curr + g.Len

	for s.curr < end {
		f := new(field)
		f.Type = btoa(s.advance(FieldTagLen))
		f.Len = s.scanLen()
		f.Value = s.advance(f.Len)

		fields = append(fields, f)
	}

	g.Fields = fields
}

// utilities

// step size means how many bytes to capture and advance
func (s *Scanner) advance(stepSize int) []byte {
	ret := s.src[s.curr : s.curr+stepSize]
	s.curr += stepSize

	return ret
}

// step advances only 1 byte
func (s *Scanner) step() byte {
	ret := s.src[s.curr]
	s.curr++

	return ret
}

// scan length, length values are variable-width encoded
// first byte can be 81,82 or < 81
// 81 => next byte specifies the size
// 82 => next 2 bytes (big-endian) specifies the size
// else first byte is the size
func (s *Scanner) scanLen() int {
	val := 0

	if v := s.step(); v == 0x81 {

		// advance to get the size in betwen 128-255
		b := s.step()
		val = int(b)

	} else if v == 0x82 {
		b := s.advance(2)

		left := int(b[0]) << 8
		right := int(b[1])
		size := left + right
		val = size

	} else {
		val = int(v)
	}

	return val
}

// byte buffer to string
func btoa(b []byte) string {
	s := ""
	for _, v := range b {
		s += fmt.Sprintf("%02x", v)
	}
	return s
}

func printSeperateHex(val []byte) string {
	s := "["
	for _, v := range val {
		s += " " + fmt.Sprintf("%02x", v)
	}
	s += "]"
	return s
}
