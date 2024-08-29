package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
)

type clientConfig struct {
	id            int
	serverAddress string
}

type client struct {
	config clientConfig
	conn   net.Conn
}

func newClient(config clientConfig) *client {
	client := &client{
		config: config,
	}
	return client
}

func (c *client) createClientSocket() {
	conn, err := net.Dial("tcp", c.config.serverAddress)
	if err != nil {
		log.Fatalf(common.FmtLog("action", "connect",
			"result", "fail",
			"client_id", c.config.id,
			"error", err,
		))
	}
	c.conn = conn
}

func (c *client) sendBet(bet common.LocalBet) (err error) {
	c.createClientSocket()
	defer func() {
		closeErr := c.closeClientSocket()
		err = errors.Join(err, closeErr)
	}()

	writer := csv.NewWriter(c.conn)
	_ = writer.Write([]string{"HELLO", strconv.Itoa(c.config.id)})
	_ = writer.Write(bet.ToRecord())
	writer.Flush()
	err = writer.Error()
	if err != nil {
		return
	}

	reader := csv.NewReader(c.conn)
	response, err := reader.Read()
	if err != nil {
		return
	}

	code := response[0]
	switch code {
	case "OK":
		return
	default:
		return fmt.Errorf("server didn't send OK: %v", code)
	}
}

func (c *client) closeClientSocket() error {
	err := c.conn.Close()
	if err != nil {
		log.Error(common.FmtLog("action", "close_connection",
			"result", "fail",
			"error", err,
		))
		return err
	}
	return nil
}
