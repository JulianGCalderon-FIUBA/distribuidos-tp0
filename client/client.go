package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net"

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
	_ = writer.Write(common.Hello{AgencyId: c.config.id}.ToRecord())
	_ = writer.Write(bet.ToRecord())
	writer.Flush()
	err = writer.Error()
	if err != nil {
		return
	}

	reader := csv.NewReader(c.conn)
	okRecord, err := reader.Read()
	if err != nil {
		return
	}
	_, err = common.OkFromRecord(okRecord)
	if err != nil {
		return fmt.Errorf("server didn't send ok")
	}

	return
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
