package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
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

	_ = c.writer.Write(common.Hello{AgencyId: c.config.id}.ToRecord())
	c.writer.Flush()
	err = c.writer.Error()
	if err != nil {
		return fmt.Errorf("failed to send ok: %w", err)
	}

	return nil
}

func (c *client) sendBets(ctx context.Context, bets []common.LocalBet) (err error) {
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

	for len(bets) > 0 {
		currentBatchEnd := min(len(bets), c.config.batchSize)

		var batch []common.LocalBet
		bets, batch = bets[currentBatchEnd:], bets[:currentBatchEnd]

		err := c.sendBatch(batch)
		if err != nil {
			log.Error(common.FmtLog(
				"action", "send_batch",
				"result", "fail",
				"error", err,
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

func (c *client) sendBatch(bets []common.LocalBet) (err error) {
	_ = c.writer.Write(common.Batch{BatchSize: len(bets)}.ToRecord())
	for _, bet := range bets {
		_ = c.writer.Write(bet.ToRecord())
	}
	c.writer.Flush()
	err = c.writer.Error()
	if err != nil {
		return
	}

	okRecord, err := c.reader.Read()
	if err != nil {
		return
	}
	_, err = common.OkFromRecord(okRecord)
	if err != nil {
		return fmt.Errorf("server didn't send ok")
	}

	log.Info(common.FmtLog(
		"action", "send_batch",
		"result", "success",
		"batchSize", len(bets),
	))

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
