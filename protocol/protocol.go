package protocol

import (
	"encoding/csv"
	"fmt"
	"time"
)

type MessageCode string

const (
	HelloCode   MessageCode = "HELLO"
	BatchCode   MessageCode = "BATCH"
	BetCode     MessageCode = "BET"
	OkCode      MessageCode = "OK"
	ErrCode     MessageCode = "ERR"
	FinishCode  MessageCode = "FINISH"
	WinnersCode MessageCode = "WINNERS"
)

type Message interface {
	Code() MessageCode
}

func Send(m Message, w *csv.Writer) error {
	rawData := Serialize(m)
	data := append([]string{string(m.Code())}, rawData...)

	_ = w.Write(data)
	w.Flush()
	return w.Error()
}

func ReceiveAny(r *csv.Reader) (m Message, err error) {
	record, err := r.Read()
	if err != nil {
		return
	}

	switch MessageCode(record[0]) {
	case HelloCode:
		return Deserialize[HelloMessage](record[1:])
	case BatchCode:
		return Deserialize[BatchMessage](record[1:])
	case BetCode:
		return Deserialize[BetMessage](record[1:])
	case OkCode:
		return Deserialize[OkMessage](record[1:])
	case ErrCode:
		return Deserialize[ErrMessage](record[1:])
	case FinishCode:
		return Deserialize[FinishMessage](record[1:])
	default:
		return m, fmt.Errorf("invalid MessageCode")
	}
}

func Receive[M Message](r *csv.Reader) (M, error) {
	var m M

	record, err := r.Read()
	if err != nil {
		return m, err
	}

	if record[0] != string(m.Code()) {
		return m, fmt.Errorf("expected code %v, got %v", m.Code(), record[0])
	}

	return Deserialize[M](record[1:])
}

type HelloMessage struct {
	AgencyId int
}

type BatchMessage struct {
	BatchSize int
}

type BetMessage struct {
	FirstName string
	LastName  string
	Document  int
	Birthdate time.Time
	Number    int
}

type OkMessage struct{}

type ErrMessage struct{}

type FinishMessage struct{}

type WinnersMessage []int

func (m BatchMessage) Code() MessageCode {
	return BatchCode
}

func (m HelloMessage) Code() MessageCode {
	return HelloCode
}

func (m BetMessage) Code() MessageCode {
	return BetCode
}

func (m OkMessage) Code() MessageCode {
	return OkCode
}

func (m ErrMessage) Code() MessageCode {
	return ErrCode
}

func (m FinishMessage) Code() MessageCode {
	return FinishCode
}

func (m WinnersMessage) Code() MessageCode {
	return WinnersCode
}
