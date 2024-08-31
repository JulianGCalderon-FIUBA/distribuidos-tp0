package lottery_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/server/lottery"
)

func TestStorage(t *testing.T) {
	var file bytes.Buffer

	input_bets := []lottery.Bet{
		{
			Agency:    1,
			FirstName: "laura",
			LastName:  "lopez",
			Document:  40000001,
			Birthdate: time.Date(2001, time.May, 1, 0, 0, 0, 0, time.UTC),
			Number:    1,
		},
		{
			Agency:    2,
			FirstName: "juan",
			LastName:  "jerez",
			Document:  40000002,
			Birthdate: time.Date(2002, time.May, 2, 0, 0, 0, 0, time.UTC),
			Number:    2,
		},
		{
			Agency:    3,
			FirstName: "mateo",
			LastName:  "melasco",
			Document:  40000003,
			Birthdate: time.Date(2003, time.May, 3, 0, 0, 0, 0, time.UTC),
			Number:    3,
		},
	}

	_ = lottery.StoreBetsIn(&file, input_bets[:2])
	_ = lottery.StoreBetsIn(&file, input_bets[2:])
	output_bets, _ := lottery.LoadBetsFrom(&file)

	if !reflect.DeepEqual(input_bets, output_bets) {
		t.Fatalf("expected %v, but got %v", input_bets, output_bets)
	}
}
