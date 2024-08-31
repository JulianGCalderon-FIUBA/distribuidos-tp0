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
	agencyId int
	conn     *net.TCPConn
	reader   *csv.Reader
	writer   *csv.Writer
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

		for agencyIndex, handler := range s.connections {
			if handler.conn == nil {
				continue
			}
			agencyId := agencyIndex + 1

			batchSize, err := handler.receiveBatch()
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			if err != nil {
				_ = closeConnection(s.connections[agencyIndex].conn)
				s.connections[agencyIndex].conn = nil

				log.Error(common.FmtLog("action", "receive_batch",
					"result", "fail",
					"agency_id", agencyId,
					"error", err,
				))
			} else {
				log.Info(common.FmtLog("action", "receive_batch",
					"result", "success",
					"agency_id", agencyId,
					"batch_size", batchSize,
				))
			}
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

func (h handler) receiveBatch() (int, error) {
	err := h.conn.SetDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		return 0, err
	}
	batch, err := protocol.Receive[protocol.BatchMessage](h.reader)
	if err != nil {
		return 0, err
	}

	err = h.conn.SetDeadline(time.Time{})
	if err != nil {
		return 0, err
	}

	bets := make([]lottery.Bet, 0, batch.BatchSize)

	for i := 0; i < batch.BatchSize; i++ {
		betMessage, err := protocol.Receive[protocol.BetMessage](h.reader)
		if err != nil {
			return 0, fmt.Errorf("failed to parse bet: %w", err)
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

	err = lottery.StoreBets(bets)
	if err != nil {
		storeErr := fmt.Errorf("failed to store bets: %w", err)

		sendErr := protocol.Send(protocol.ErrMessage{}, h.writer)
		if sendErr != nil {
			sendErr = fmt.Errorf("failed to send err message: %w", err)
		}

		return 0, errors.Join(storeErr, sendErr)
	}

	err = protocol.Send(protocol.OkMessage{}, h.writer)
	if err != nil {
		return 0, fmt.Errorf("failed to send ok: %w", err)
	}

	return batch.BatchSize, nil
}

func (s *server) hasConnection() bool {
	for _, h := range s.connections {
		if h.conn != nil {
			return true
		}
	}
	return false
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
