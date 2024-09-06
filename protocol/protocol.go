package protocol

import (
	"fmt"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/safeio"
)

// This module defines the discriminant (type) and structure of each
// message that is sent over the network.
//
// To serialize this structures, I made my own generic serialization
// library. It uses reflection (runtime data structure inspection)
// to figure out how to serialize (and deserialize) each type into a
// list of strings.  This is how generic serialization packages like
// `encondig/json` work.  In theory it could serialize any type, but in
// practice I only implemented the necessary features for my particular
// protocol.
//
// This approach was not necessary, manually implementing the
// serialization methods would probably have been faster (and more
// performant). I did it this way as a personal challenge, as I've been
// wanting to try out reflection for a long time.

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

// Serializes a message as a list of strings and writes it to the writter.
// If the message contains a comma, the field will be surrounded by
// double quotes.
func Send(m Message, w *safeio.Writer) {
	rawData := Serialize(m)
	data := append([]string{string(m.Code())}, rawData...)

	w.Write(data)
}

// Like `Send`, but flushes the buffer afterwards.
func SendFlush(m Message, w *safeio.Writer) error {
	Send(m, w)
	return Flush(w)
}

// Flushes the buffer and returns any errors encountered
func Flush(w *safeio.Writer) error {
	return w.Flush()
}

func ReceiveAny(r *safeio.Reader) (m Message, err error) {
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

func Receive[M Message](r *safeio.Reader) (M, error) {
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
