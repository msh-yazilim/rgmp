package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mselh/rgmp/scan"
)

func main() {

	// reads from stdin
	fd := os.Stdin
	buf, err := io.ReadAll(fd)
	if err != nil {
		panic("no std in feed provided")
	}

	s := scan.NewScanner(buf)
	req, err := s.Scan()
	if err != nil {
		panic(err)
	}

	fmt.Println(req.String())
	// parse field messages
}
