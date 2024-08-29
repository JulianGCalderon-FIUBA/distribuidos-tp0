package common

import (
	"fmt"
	"strconv"
)

type Hello struct {
	AgencyId  int
	BatchSize int
}

func (h Hello) ToRecord() []string {
	return []string{
		"HELLO",
		strconv.Itoa(h.AgencyId),
		strconv.Itoa(h.BatchSize),
	}
}

func HelloFromRecord(record []string) (hello Hello, err error) {
	if len(record) != 3 {
		err = fmt.Errorf("record should contains 3 fields")
		return
	}

	if record[0] != "HELLO" {
		err = fmt.Errorf("first record should be HELLO")
		return
	}

	agencyId, err := strconv.Atoi(record[1])
	if err != nil {
		return
	}
	hello.AgencyId = agencyId

	batchSize, err := strconv.Atoi(record[2])
	if err != nil {
		return
	}
	hello.BatchSize = batchSize

	return
}

type Ok struct{}

func (o Ok) ToRecord() []string {
	return []string{
		"OK",
	}
}

func OkFromRecord(record []string) (ok Ok, err error) {
	if len(record) != 1 {
		err = fmt.Errorf("record should contains 2 fields")
		return
	}

	if record[0] != "OK" {
		err = fmt.Errorf("first record should be OK")
		return
	}

	return
}

type Err struct{}

func (o Err) ToRecord() []string {
	return []string{
		"ERR",
	}
}
