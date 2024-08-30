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
	_ = ctx

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

		for agencyId, handler := range s.connections {
			if handler.conn == nil {
				continue
			}

			err := handler.receiveBatch()
			if err != nil {
				log.Error(common.FmtLog("action", "apuesta_recibida",
					"result", "fail",
					"agency_id", agencyId,
					"error", err,
				))
			}
		}
	}
}

func (s *server) acceptConnection() error {
	err := s.listener.SetDeadline(time.Now().Add(500 * time.Millisecond))
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

	helloRecord, err := handler.reader.Read()
	if err != nil {
		closeErr := closeConnection(conn)
		return errors.Join(err, closeErr)
	}
	hello, err := common.HelloFromRecord(helloRecord)
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

func (h handler) receiveBatch() (err error) {
	err = h.conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		return
	}

	batchRecord, err := h.reader.Read()
	if errors.Is(err, os.ErrDeadlineExceeded) {
		err = nil
		return
	}
	if err != nil {
		return
	}

	err = h.conn.SetDeadline(time.Time{})
	if err != nil {
		return
	}

	batch, err := common.BatchFromRecord(batchRecord)
	if err != nil {
		return fmt.Errorf("failed to parse batch: %w", err)
	}

	bets := make([]lottery.Bet, 0, batch.BatchSize)

	for i := 0; i < batch.BatchSize; i++ {
		var localBetRecord []string
		localBetRecord, err = h.reader.Read()
		if err != nil {
			return
		}
		localBet, err := common.LocalBetFromRecord(localBetRecord)
		if err != nil {
			return fmt.Errorf("failed to parse bet: %w", err)
		}
		bet := lottery.Bet{
			Agency:   h.agencyId,
			LocalBet: localBet,
		}

		bets = append(bets, bet)
	}

	err = lottery.StoreBets(bets)
	if err != nil {
		storeErr := fmt.Errorf("failed to store bets: %w", err)

		_ = h.writer.Write(common.Err{}.ToRecord())
		h.writer.Flush()
		sendErr := h.writer.Error()
		if sendErr != nil {
			sendErr = fmt.Errorf("failed to send err message: %w", err)
		}

		return errors.Join(storeErr, sendErr)
	}

	_ = h.writer.Write(common.Ok{}.ToRecord())
	h.writer.Flush()
	err = h.writer.Error()
	if err != nil {
		return fmt.Errorf("failed to send ok: %w", err)
	}

	log.Info(common.FmtLog("action", "apuesta_recibida",
		"result", "success",
		"agency_id", h.agencyId,
		"cantidad", batch.BatchSize,
	))

	return
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
