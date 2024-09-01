package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/server/lottery"
)

const MAX_AGENCIES = 5

type server struct {
	listener    *net.TCPListener
	connections [MAX_AGENCIES]handler
}

type handler struct {
	agencyId  int
	conn      *net.TCPConn
	reader    *csv.Reader
	writer    *csv.Writer
	finalized bool
}

func newServer(port int, listenBacklog int) (*server, error) {
	_ = listenBacklog
	address_string := fmt.Sprintf("0.0.0.0:%v", port)
	address, err := net.ResolveTCPAddr("tcp", address_string)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return nil, err
	}

	return &server{
		listener: listener,
	}, nil
}

func (s *server) run(ctx context.Context) (err error) {
	defer func() {
		errs := make([]error, 0, len(s.connections)+2)
		if err != nil {
			errs = append(errs, err)
		}

		err = closeListener(s.listener)
		if err != nil {
			errs = append(errs, err)
		}

		for _, handler := range s.connections {
			if handler.conn == nil {
				continue
			}
			err := closeConnection(handler.conn)
			if err != nil {
				errs = append(errs, err)
			}
		}

		err = errors.Join(errs...)
	}()

	for {
		err = s.acceptConnection()
		if err != nil {
			log.Info(common.FmtLog("action", "connected",
				"result", "fail",
				"error", err,
			))
			return
		}

		for i := range s.connections {
			handler := &s.connections[i]

			if handler.conn == nil || handler.finalized {
				continue
			}
			s.handleClient(handler)
		}

		if s.hasFinalized() {
			log.Info(common.FmtLog(
				"action", "sorteo",
				"result", "success",
			))

			winners, err := getWinners()
			if err != nil {
				return err
			}

			for agencyIdx := range s.connections {
				h := &s.connections[agencyIdx]

				agencyWinners := winners[agencyIdx+1]
				err := h.sendWinners(agencyWinners)
				if err != nil {
					log.Error(common.FmtLog(
						"action", "send_winners",
						"result", "fail",
						"error", err,
					))
				}
			}

			return nil
		}

		if !s.hasConnection() {
			select {
			case <-ctx.Done():
				log.Info(common.FmtLog(
					"action", "shutdown",
				))
				return nil
			default:
			}
		}

	}
}

func (s *server) handleClient(h *handler) {
	message, err := h.receiveRequest()
	if err != nil {
		s.closeHandler(h.agencyId)
		log.Error(common.FmtLog("action", "receive_request",
			"result", "fail",
			"agency_id", h.agencyId,
			"error", err,
		))
	}
	if message == nil {
		return
	}

	switch message := message.(type) {
	case protocol.BatchMessage:
		err = h.receiveBatch(message.BatchSize)
		if err != nil {
			s.closeHandler(h.agencyId)

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
		h.finalized = true
	}
}

func (s *server) acceptConnection() error {
	err := s.listener.SetDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		return err
	}

	conn, err := s.listener.AcceptTCP()
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return nil
	}

	handler := handler{
		conn:   conn,
		reader: csv.NewReader(conn),
		writer: csv.NewWriter(conn),
	}
	handler.reader.FieldsPerRecord = -1

	hello, err := protocol.Receive[protocol.HelloMessage](handler.reader)
	if err != nil {
		closeErr := closeConnection(conn)
		return errors.Join(err, closeErr)
	}
	handler.agencyId = hello.AgencyId
	agencyIndex := hello.AgencyId - 1

	if s.connections[agencyIndex].conn != nil {
		err := closeConnection(conn)
		log.Error(common.FmtLog("action", "connect",
			"result", "fail",
			"agency_id", hello.AgencyId,
			"error", "already connected",
		))

		return err
	}
	s.connections[agencyIndex] = handler

	log.Info(common.FmtLog("action", "connect",
		"result", "success",
		"agency_id", hello.AgencyId,
	))

	return nil
}

func (h *handler) receiveRequest() (protocol.Message, error) {
	err := h.conn.SetDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		return nil, err
	}
	message, err := protocol.ReceiveAny(h.reader)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.conn.SetDeadline(time.Time{})
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (h handler) receiveBatch(batchSize int) error {
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

	err := lottery.StoreBets(bets)
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

func (h *handler) sendWinners(winners []int) error {
	return protocol.Send(protocol.WinnersMessage(winners), h.writer)
}

func (s *server) hasConnection() bool {
	for _, h := range s.connections {
		if h.conn != nil && !h.finalized {
			return true
		}
	}
	return false
}

func (s *server) hasFinalized() bool {
	for _, h := range s.connections {
		if !h.finalized {
			return false
		}
	}
	return true
}

func (s *server) closeHandler(agencyId int) {
	_ = closeConnection(s.connections[agencyId-1].conn)
	s.connections[agencyId-1].conn = nil
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

func closeListener(listener net.Listener) error {
	err := listener.Close()
	if err != nil {
		log.Error(common.FmtLog("action", "close_listener",
			"result", "fail",
			"error", err,
		))
		return err
	}
	return nil
}

func getWinners() (map[int][]int, error) {
	allBets, err := lottery.LoadBets()
	if err != nil {
		return nil, err
	}

	winners := make(map[int][]int)

	for _, bet := range allBets {
		if bet.HasWon() {
			winners[bet.Agency] = append(winners[bet.Agency], bet.Document)
		}
	}

	return winners, nil
}
