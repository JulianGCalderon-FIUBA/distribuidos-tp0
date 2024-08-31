package protocol_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
)

func TestReflect(t *testing.T) {
	messages := []any{
		protocol.HelloMessage{83},
		protocol.BatchMessage{83},
		protocol.BetMessage{
			"Laura",
			"Lopez",
			44160273,
			time.Date(2002, time.May, 16, 0, 0, 0, 0, time.UTC),
			83,
		},
		protocol.OkMessage{},
		protocol.WinnersMessage{1, 2, 3},
		protocol.WinnersMessage{},
	}

	for _, message := range messages {
		serialized := protocol.Serialize(message)

		var deserialized any
		var err error

		switch message.(type) {
		case protocol.HelloMessage:
			deserialized, err = protocol.Deserialize[protocol.HelloMessage](serialized)
		case protocol.BatchMessage:
			deserialized, err = protocol.Deserialize[protocol.BatchMessage](serialized)
		case protocol.BetMessage:
			deserialized, err = protocol.Deserialize[protocol.BetMessage](serialized)
		case protocol.OkMessage:
			deserialized, err = protocol.Deserialize[protocol.OkMessage](serialized)
		case protocol.WinnersMessage:
			deserialized, err = protocol.Deserialize[protocol.WinnersMessage](serialized)
		}

		if err != nil {
			t.Fatalf("%v", err)
		}

		if !reflect.DeepEqual(deserialized, message) {
			t.Fatalf("%#v, %#v", deserialized, message)
		}
	}
}
