package main

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/mycsv"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
)

type clientConfig struct {
	id            int
	batchSize     int
	serverAddress string
	loopPeriod    time.Duration
}

type client struct {
	config     clientConfig
	conn       *net.TCPConn
	connReader *mycsv.Reader
	connWriter *mycsv.Writer
	betsReader *mycsv.Reader
}

func newClient(config clientConfig, betsReader *mycsv.Reader) *client {
	client := &client{
		config:     config,
		betsReader: betsReader,
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
	c.connReader = mycsv.NewReader(conn)
	c.connWriter = mycsv.NewWriter(conn)

	err = protocol.SendFlush(protocol.HelloMessage{AgencyId: c.config.id}, c.connWriter)
	if err != nil {
		closeErr := closeSocket(c.conn)
		return errors.Join(err, closeErr)
	}

	return nil
}

func (c *client) run(ctx context.Context) (err error) {
	err = c.createClientSocket()
	if err != nil {
		return err
	}
	closer := common.SpawnCloser(ctx, c.conn, closeSocket)
	defer func() {
		closeErr := closer.Close()
		err = errors.Join(err, closeErr)
	}()

	for {
		batch, err := c.readBatch()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		err = c.sendBatch(batch)
		if err != nil {
			log.Error(common.FmtLog("send_batch", err))
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return err
			}
		} else {
			log.Info(common.FmtLog("send_batch", nil,
				"batchSize", len(batch),
			))
		}

		select {
		case <-ctx.Done():
			return net.ErrClosed
		case <-time.After(c.config.loopPeriod):
		}
	}

	err = protocol.SendFlush(protocol.FinishMessage{}, c.connWriter)
	if err != nil {
		return err
	}

	winners, err := protocol.Receive[protocol.WinnersMessage](c.connReader)
	if err != nil {
		return err
	} else {
		log.Info(common.FmtLog("consulta_ganadores", nil,
			"cant_ganadores", len(winners),
		))
	}

	return nil
}

func (c *client) sendBatch(bets []protocol.BetMessage) error {
	err := protocol.SendFlush(protocol.BatchMessage{BatchSize: len(bets)}, c.connWriter)
	if err != nil {
		return err
	}

	for _, bet := range bets {
		protocol.Send(bet, c.connWriter)
	}
	err = protocol.Flush(c.connWriter)
	if err != nil {
		return err
	}

	_, err = protocol.Receive[protocol.OkMessage](c.connReader)
	if err != nil {
		return err
	}

	return nil
}

// Reads batch from agency data, with up to `c.config.batchSize` bets
func (c *client) readBatch() ([]protocol.BetMessage, error) {
	batch := make([]protocol.BetMessage, 0, c.config.batchSize)

	for i := 0; i < c.config.batchSize; i++ {
		betRecord, err := c.betsReader.Read()
		if errors.Is(err, io.EOF) {
			if len(batch) > 0 {
				return batch, nil
			} else {
				return nil, io.EOF
			}
		}
		if err != nil {
			return nil, err
		}

		bet, err := protocol.Deserialize[protocol.BetMessage](betRecord)
		if err != nil {
			return nil, err
		}

		batch = append(batch, bet)
	}

	return batch, nil
}

func closeSocket(c *net.TCPConn) error {
	err := c.Close()
	if err != nil {
		log.Error(common.FmtLog("close_connection", err))
		return err
	}
	return nil
}
