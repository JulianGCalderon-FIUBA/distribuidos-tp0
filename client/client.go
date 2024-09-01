package main

import (
	"bufio"
	"context"
	"errors"
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

func (c *client) startClientLoop(ctx context.Context) {
	for msgID := 1; msgID <= c.config.loopAmount; msgID++ {
		err := c.sendMessage(msgID)
		if err != nil {
			return
		}

		// Wait a time between sending one message and the next one
		sleep_timer := time.After(c.config.loopPeriod)

		select {
		case <-ctx.Done():
			log.Infof(common.FmtLog("action", "shutdown", "result", "success"))
			return

		case <-sleep_timer:
		}

	}

	log.Infof(common.FmtLog("action", "loop_finished",
		"result", "success",
		"client_id", c.config.id,
	))
}

func (c *client) sendMessage(msgID int) (err error) {
	c.createClientSocket()
	defer func() {
		closeErr := c.closeClientSocket()
		err = errors.Join(err, closeErr)
	}()

	// TODO: Modify the send to avoid short-write
	fmt.Fprintf(
		c.conn,
		"[CLIENT %v] Message N°%v\n",
		c.config.id,
		msgID,
	)
	msg, err := bufio.NewReader(c.conn).ReadString('\n')

	if err != nil {
		log.Errorf(common.FmtLog("action", "receive_message",
			"result", "fail",
			"client_id", c.config.id,
			"error", err,
		))
	} else {
		log.Infof(common.FmtLog("action", "receive_message",
			"result", "success",
			"client_id", c.config.id,
			"msg", msg,
		))

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

	log.Info(common.FmtLog("action", "close_connection",
		"result", "success",
	))
	return nil
}
