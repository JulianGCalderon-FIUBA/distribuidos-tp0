package protocol_test

import (
	"testing"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
)

func TestReflect(t *testing.T) {
	messages := []protocol.Message{
		protocol.HelloMessage{83},
		protocol.BatchMessage{83},
		protocol.BetMessage{
			"Laura",
			"Lopez",
			44160273,
			time.Date(2002, time.May, 16, 0, 0, 0, 0, time.UTC),
			83,
		},
	}

	for _, message := range messages {
		serialized := protocol.Serialize(message)

		var deserialized protocol.Message
		var err error

		switch message.(type) {
		case protocol.HelloMessage:
			deserialized, err = protocol.Deserialize[protocol.HelloMessage](serialized)
		case protocol.BatchMessage:
			deserialized, err = protocol.Deserialize[protocol.BatchMessage](serialized)
		case protocol.BetMessage:
			deserialized, err = protocol.Deserialize[protocol.BetMessage](serialized)
		}

		if err != nil {
			t.Fatalf("%v", err)
		}

		if deserialized != message {
			t.Fatalf("%#v, %#v", deserialized, message)
		}
	}
}
