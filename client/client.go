package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
)

type clientConfig struct {
	id            int
	serverAddress string
	loopAmount    int
	loopPeriod    time.Duration
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

func (c *client) startClientLoop() {
	for msgID := 1; msgID <= c.config.loopAmount; msgID++ {
		c.createClientSocket()

		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.id,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			log.Errorf(common.FmtLog("action", "receive_message",
				"result", "fail",
				"client_id", c.config.id,
				"error", err,
			))
			return
		}

		log.Infof(common.FmtLog("action", "receive_message",
			"result", "success",
			"client_id", c.config.id,
			"msg", msg,
		))

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.loopPeriod)

	}

	log.Infof(common.FmtLog("action", "loop_finished",
		"result", "success",
		"client_id", c.config.id,
	))
}
