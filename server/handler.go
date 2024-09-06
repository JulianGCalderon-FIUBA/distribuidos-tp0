package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/mycsv"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/server/lottery"
)

type handler struct {
	agencyId int
	conn     net.Conn
	reader   *mycsv.Reader
	writer   *mycsv.Writer
	server   *server
}

func createHandler(s *server, conn net.Conn) (*handler, error) {
	reader := mycsv.NewReader(conn)

	hello, err := protocol.Receive[protocol.HelloMessage](reader)
	if err != nil {
		return nil, err
	}

	return &handler{
		agencyId: hello.AgencyId,
		conn:     conn,
		reader:   reader,
		writer:   mycsv.NewWriter(conn),
		server:   s,
	}, nil
}

func (h *handler) run(ctx context.Context) (err error) {
	closer := common.SpawnCloser(ctx, h.conn, closeConnection)
	defer func() {
		closeErr := closer.Close()
		err = errors.Join(err, closeErr)
	}()

	for {
		var message protocol.Message
		message, err := protocol.ReceiveAny(h.reader)
		if err != nil {
			return err
		}

		switch message := message.(type) {
		case protocol.BatchMessage:
			err = h.receiveBatch(message.BatchSize)
			if err != nil {
				log.Error(common.FmtLog("receive_batch", err,
					"agency_id", h.agencyId,
				))
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					return err
				}
			} else {
				log.Info(common.FmtLog("receive_batch", nil,
					"agency_id", h.agencyId,
					"batch_size", message.BatchSize,
				))
			}
		case protocol.FinishMessage:
			h.server.lotteryFinish.Done()

			log.Info(common.FmtLog("receive_finish", nil,
				"agency_id", h.agencyId,
			))

			err := h.sendWinners(ctx)
			if err != nil {
				return err
			}

			return nil
		}
	}
}

func (h *handler) sendWinners(ctx context.Context) error {
	lotteryFinish := make(chan struct{})
	go func() {
		h.server.lotteryFinish.Wait()
		lotteryFinish <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return net.ErrClosed
	case <-lotteryFinish:
		if h.agencyId == 1 {
			log.Info(common.FmtLog("sorteo", nil))
		}

		winners, err := h.server.getWinners()
		if err != nil {
			return err
		}

		err = protocol.SendFlush(protocol.WinnersMessage(winners[h.agencyId]), h.writer)
		if err != nil {
			return err
		}

		return nil
	}
}

func (h *handler) receiveBatch(batchSize int) error {
	bets := make([]lottery.Bet, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		betMessage, err := protocol.Receive[protocol.BetMessage](h.reader)
		if err != nil {
			return fmt.Errorf("failed to parse bet: %w", err)
		}

		bet := lottery.Bet{
			Agency:    h.agencyId,
			FirstName: betMessage.FirstName,
			LastName:  betMessage.LastName,
			Document:  betMessage.Document,
			Birthdate: betMessage.Birthdate,
			Number:    betMessage.Number,
		}

		bets = append(bets, bet)
	}

	h.server.storageLock.Lock()
	storeErr := lottery.StoreBets(bets)
	h.server.storageLock.Unlock()
	if storeErr != nil {
		storeErr = fmt.Errorf("failed to store bets: %w", storeErr)
		sendErr := protocol.SendFlush(protocol.ErrMessage{}, h.writer)
		return errors.Join(storeErr, sendErr)
	}

	err := protocol.SendFlush(protocol.OkMessage{}, h.writer)
	if err != nil {
		return err
	}

	return nil
}

func closeConnection(conn net.Conn) error {
	err := conn.Close()
	if err != nil {
		log.Error(common.FmtLog("close_connection", err))
		return err
	}
	return nil
}
