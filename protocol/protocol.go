package protocol

import (
	"encoding/csv"
	"fmt"
	"time"
)

type MessageCode string

const (
	HelloCode  MessageCode = "HELLO"
	BatchCode  MessageCode = "BATCH"
	BetCode    MessageCode = "BET"
	OkCode     MessageCode = "OK"
	ErrCode    MessageCode = "ERR"
	FinishCode MessageCode = "FINISH"
)

type Message interface {
	Code() MessageCode
}

func Send(m Message, w *csv.Writer) error {
	return w.Write(Serialize(m))
}

func ReceiveAny(r *csv.Reader) (m Message, err error) {
	record, err := r.Read()
	if err != nil {
		return
	}

	switch MessageCode(record[0]) {
	case HelloCode:
		return Deserialize[HelloMessage](record)
	case BatchCode:
		return Deserialize[BatchMessage](record)
	case BetCode:
		return Deserialize[BetMessage](record)
	case OkCode:
		return Deserialize[OkMessage](record)
	case ErrCode:
		return Deserialize[ErrMessage](record)
	case FinishCode:
		return Deserialize[FinishMessage](record)
	default:
		return m, fmt.Errorf("invalid MessageCode")
	}
}

func Receive[M Message](r *csv.Reader) (m M, err error) {
	record, err := r.Read()
	if err != nil {
		return
	}

	return Deserialize[M](record)
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
