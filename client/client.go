package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
)

type clientConfig struct {
	id            int
	batchSize     int
	serverAddress string
}

type client struct {
	config clientConfig
	conn   *net.TCPConn
	reader *csv.Reader
	writer *csv.Writer
}

func newClient(config clientConfig) *client {
	client := &client{
		config: config,
	}
	return client
}

func (c *client) createClientSocket() error {
	raddr, err := net.ResolveTCPAddr("tcp", c.config.serverAddress)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return err
	}

	c.conn = conn
	c.reader = csv.NewReader(conn)
	c.writer = csv.NewWriter(conn)

	err = protocol.Send(protocol.HelloMessage{AgencyId: c.config.id}, c.writer)
	if err != nil {
		closeErr := c.closeClientSocket()
		return errors.Join(err, closeErr)
	}

	return nil
}

func (c *client) sendBets(ctx context.Context, bets []protocol.BetMessage) (err error) {
	err = c.createClientSocket()
	if err != nil {
		log.Fatalf(common.FmtLog("action", "connect",
			"result", "fail",
			"error", err,
		))
	}
	defer func() {
		closeErr := c.closeClientSocket()
		err = errors.Join(err, closeErr)
	}()

	for _, batch := range batchBets(bets, c.config.batchSize) {
		err := c.sendBatch(batch)
		if err != nil {
			log.Error(common.FmtLog(
				"action", "send_batch",
				"result", "fail",
				"error", err,
			))
		} else {
			log.Info(common.FmtLog(
				"action", "send_batch",
				"result", "success",
				"batchSize", len(bets),
			))
		}

		select {
		case <-ctx.Done():
			log.Info(common.FmtLog(
				"action", "shutdown",
			))
			return nil
		default:
		}
	}

	log.Info(common.FmtLog(
		"action", "finished",
	))

	return nil
}

func (c *client) sendBatch(bets []protocol.BetMessage) error {
	err := protocol.Send(protocol.BatchMessage{BatchSize: len(bets)}, c.writer)
	if err != nil {
		return err
	}

	for _, bet := range bets {
		err := protocol.Send(bet, c.writer)
		if err != nil {
			return err
		}
	}

	_, err = protocol.Receive[protocol.OkMessage](c.reader)
	if err != nil {
		return fmt.Errorf("server didn't send ok")
	}

	return nil
}

// Creates an iterator of batches of `bets`, each with max `batchSize` elements
func batchBets(bets []protocol.BetMessage, batchSize int) [][]protocol.BetMessage {
	batches := make([][]protocol.BetMessage, 0)

	for len(bets) > 0 {
		currentBatchEnd := min(len(bets), batchSize)

		var batch []protocol.BetMessage
		bets, batch = bets[currentBatchEnd:], bets[:currentBatchEnd]

		batches = append(batches, batch)
	}

	return batches
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
