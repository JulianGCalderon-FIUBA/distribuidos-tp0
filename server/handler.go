package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/server/lottery"
)

type handler struct {
	agencyId int
	conn     *net.TCPConn
	reader   *csv.Reader
	writer   *csv.Writer
	server   *server
}

func createHandler(s *server, conn *net.TCPConn) (*handler, error) {
	reader := csv.NewReader(conn)
	reader.FieldsPerRecord = -1

	hello, err := protocol.Receive[protocol.HelloMessage](reader)
	if err != nil {
		return nil, err
	}

	return &handler{
		agencyId: hello.AgencyId,
		conn:     conn,
		reader:   reader,
		writer:   csv.NewWriter(conn),
		server:   s,
	}, nil
}

func (h *handler) run(ctx context.Context) (err error) {
	defer func() {
		closeErr := closeConnection(h.conn)
		err = errors.Join(err, closeErr)
	}()

	for {

		select {
		case <-ctx.Done():
			log.Info(common.FmtLog(
				"action", "shutdown_connection",
				"result", "success",
				"agency_id", h.agencyId,
			))
			return nil
		default:
		}

		var message protocol.Message
		message, err := protocol.ReceiveAny(h.reader)
		if err != nil {
			return err
		}

		switch message := message.(type) {
		case protocol.BatchMessage:
			err = h.receiveBatch(message.BatchSize)
			if err != nil {
				log.Error(common.FmtLog("action", "receive_batch",
					"result", "fail",
					"agency_id", h.agencyId,
					"error", err,
				))
			} else {
				log.Info(common.FmtLog("action", "receive_batch",
					"result", "success",
					"agency_id", h.agencyId,
					"batch_size", message.BatchSize,
				))
			}
		case protocol.FinishMessage:
			log.Info(common.FmtLog("action", "receive_finish",
				"result", "success",
				"agency_id", h.agencyId,
			))
			h.server.lotteryFinish.Done()

			err := h.sendWinners(ctx)
			if err != nil {
				return err
			}
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
		log.Info(common.FmtLog(
			"action", "shutdown_connection",
			"result", "success",
			"agency_id", h.agencyId,
		))
		return nil
	case <-lotteryFinish:
		if h.agencyId == 1 {
			log.Info(common.FmtLog(
				"action", "sorteo",
				"result", "success",
			))
		}

		winners, err := h.server.getWinners()
		if err != nil {
			return err
		}

		err = protocol.Send(protocol.WinnersMessage(winners[h.agencyId]), h.writer)
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
	err := lottery.StoreBets(bets)
	h.server.storageLock.Unlock()

	if err != nil {
		storeErr := fmt.Errorf("failed to store bets: %w", err)

		sendErr := protocol.Send(protocol.ErrMessage{}, h.writer)
		if sendErr != nil {
			sendErr = fmt.Errorf("failed to send err message: %w", err)
		}

		return errors.Join(storeErr, sendErr)
	}

	err = protocol.Send(protocol.OkMessage{}, h.writer)
	if err != nil {
		return fmt.Errorf("failed to send ok: %w", err)
	}

	return nil
}

func closeConnection(conn net.Conn) error {
	err := conn.Close()
	if err != nil {
		log.Error(common.FmtLog("action", "close_connection",
			"result", "fail",
			"error", err,
		))
		return err
	}
	return nil
}
