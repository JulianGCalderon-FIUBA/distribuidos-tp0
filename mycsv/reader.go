package mycsv

import (
	"bufio"
	"io"
	"strings"
)

type Reader struct {
	buf *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		buf: bufio.NewReader(r),
	}
}

func (r *Reader) Read() ([]string, error) {
	rawRecord, err := r.buf.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	rawRecord = rawRecord[:len(rawRecord)-1]

	// drop carriage return if exists
	if rawRecord[len(rawRecord)-1] == '\r' {
		rawRecord = rawRecord[:len(rawRecord)-1]
	}

	records := strings.Split(string(rawRecord), ",")
	return records, nil
}
