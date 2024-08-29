package lottery

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
)

const STORAGE_FILEPATH = "./bets.csv"

type Bet struct {
	Agency int
	common.LocalBet
}

func (b Bet) ToRecord() []string {
	return append(
		[]string{strconv.Itoa(b.Agency)},
		b.LocalBet.ToRecord()...,
	)
}

func BetFromRecord(record []string) (bet Bet, err error) {
	if len(record) != 6 {
		err = fmt.Errorf("record should contains 6 fields")
		return
	}

	agency, err := strconv.Atoi(record[0])
	if err != nil {
		return
	}
	bet.Agency = agency

	localBet, err := common.LocalBetFromRecord(record[1:])
	if err != nil {
		return
	}
	bet.LocalBet = localBet

	return
}

// Persist the information of each bet in the STORAGE_FILEPATH file.
// Not thread-safe/process-safe.
func StoreBets(bets []Bet) (err error) {
	file, err := os.OpenFile(STORAGE_FILEPATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer func() {
		closeErr := file.Close()
		err = errors.Join(err, closeErr)
	}()

	err = StoreBetsIn(file, bets)
	return
}

func StoreBetsIn(w io.Writer, bets []Bet) (err error) {
	writer := csv.NewWriter(w)

	for _, bet := range bets {
		err = writer.Write(bet.ToRecord())
		if err != nil {
			return
		}
	}

	writer.Flush()
	err = writer.Error()

	return
}

// Loads the information all the bets in the STORAGE_FILEPATH file.
// Not thread-safe/process-safe.
func LoadBets() (bets []Bet, err error) {
	file, err := os.Open(STORAGE_FILEPATH)
	if err != nil {
		return
	}
	defer func() {
		closeErr := file.Close()
		err = errors.Join(err, closeErr)
	}()

	bets, err = LoadBetsFrom(file)
	return
}

func LoadBetsFrom(r io.Reader) ([]Bet, error) {
	reader := csv.NewReader(r)
	bets := make([]Bet, 0)

	for {
		var row []string
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return bets, err
		}

		bet, err := BetFromRecord(row)
		if err != nil {
			return bets, err
		}

		bets = append(bets, bet)
	}

	return bets, nil
}
