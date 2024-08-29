package common

import (
	"fmt"
	"strconv"
	"time"
)

type LocalBet struct {
	FirstName string
	LastName  string
	Document  int
	Birthdate time.Time
	Number    int
}

func (b LocalBet) ToRecord() []string {
	return []string{
		b.FirstName,
		b.LastName,
		strconv.Itoa(b.Document),
		b.Birthdate.Format(time.DateOnly),
		strconv.Itoa(b.Number),
	}
}

func LocalBetFromRecord(record []string) (bet LocalBet, err error) {
	if len(record) != 5 {
		err = fmt.Errorf("record should contains 5 fields")
		return
	}

	bet.FirstName = record[0]
	bet.LastName = record[1]

	document, err := strconv.Atoi(record[2])
	if err != nil {
		return
	}
	bet.Document = document

	birthdate, err := time.Parse(time.DateOnly, record[3])
	if err != nil {
		return
	}
	bet.Birthdate = birthdate

	number, err := strconv.Atoi(record[4])
	if err != nil {
		return
	}
	bet.Number = number

	return
}
