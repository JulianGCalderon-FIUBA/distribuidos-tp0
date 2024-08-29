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
