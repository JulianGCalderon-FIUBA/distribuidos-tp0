package lottery

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/mycsv"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
)

const STORAGE_FILEPATH = "./bets.csv"
const LOTTERY_WINNER_NUMBER = 7574

type Bet struct {
	Agency    int
	FirstName string
	LastName  string
	Document  int
	Birthdate time.Time
	Number    int
}

func (b Bet) HasWon() bool {
	return b.Number == LOTTERY_WINNER_NUMBER
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
	writer := mycsv.NewWriter(w)

	for _, bet := range bets {
		writer.Write(protocol.Serialize(bet))
	}

	err = writer.Flush()

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
	reader := mycsv.NewReader(r)
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

		bet, err := protocol.Deserialize[Bet](row)
		if err != nil {
			return bets, err
		}

		bets = append(bets, bet)
	}

	return bets, nil
}
