package main

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
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
	c.reader.FieldsPerRecord = -1

	err = protocol.Send(protocol.HelloMessage{AgencyId: c.config.id}, c.writer)
	if err != nil {
		closeErr := closeSocket(c.conn)
		return errors.Join(err, closeErr)
	}

	return nil
}

func (c *client) run(ctx context.Context, bets []protocol.BetMessage) (err error) {
	err = c.createClientSocket()
	if err != nil {
		return err
	}
	closer := common.SpawnCloser(ctx, c.conn, closeSocket)
	defer func() {
		closeErr := closer.Close()
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
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return err
			}
		} else {
			log.Info(common.FmtLog(
				"action", "send_batch",
				"result", "success",
				"batchSize", len(batch),
			))
		}
	}

	err = protocol.Send(protocol.FinishMessage{}, c.writer)
	if err != nil {
		return err
	}

	winners, err := protocol.Receive[protocol.WinnersMessage](c.reader)
	if err != nil {
		return err
	} else {
		log.Info(common.FmtLog(
			"action", "consulta_ganadores",
			"result", "success",
			"cant_ganadores", len(winners),
		))
	}

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
		return err
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

func closeSocket(c *net.TCPConn) error {
	err := c.Close()
	if err != nil {
		log.Error(common.FmtLog("action", "close_connection",
			"result", "fail",
			"error", err,
		))
		return err
	}
	return nil
}
