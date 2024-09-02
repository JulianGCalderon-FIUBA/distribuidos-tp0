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

type server struct {
	listener *net.TCPListener
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
		closeErr := closeListener(s.listener)
		err = errors.Join(err, closeErr)
	}()

	for {
		var conn net.Conn = nil

		log.Info(common.FmtLog("action", "accept_connections",
			"result", "in_progress",
		))

		// We loop until a connection is accepted, or the context finalizes.
		for {
			select {
			case <-ctx.Done():
				// If the context finalizes, just return
				log.Infof(common.FmtLog("action", "shutdown", "result", "success"))
				return
			default:
			}

			err = s.listener.SetDeadline(time.Now().Add(500 * time.Millisecond))
			if err != nil {
				return err
			}

			conn, err = s.listener.Accept()
			// If the deadline exceeded, retry
			if errors.Is(err, os.ErrDeadlineExceeded) {
				err = nil
				continue
			}
			// On any other error, just return
			if err != nil {
				return
			}

			break
		}
		addr, err := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
		if err != nil {
			return err
		}

		log.Info(common.FmtLog("action", "accept_connections",
			"result", "success",
			"ip", addr.IP,
		))

		batchSize, err := s.receiveBatch(conn)
		if err != nil {
			log.Error(common.FmtLog("action", "apuesta_recibida",
				"result", "fail",
				"cantidad", batchSize,
			))
		} else {
			log.Info(common.FmtLog("action", "apuesta_recibida",
				"result", "success",
				"cantidad", batchSize,
			))
		}
	}
}

func (s *server) receiveBatch(conn net.Conn) (batchSize int, err error) {
	defer func() {
		closeErr := closeConnection(conn)
		err = errors.Join(err, closeErr)
	}()

	reader := csv.NewReader(conn)
	reader.FieldsPerRecord = -1

	helloRecord, err := reader.Read()
	if err != nil {
		return
	}
	hello, err := common.HelloFromRecord(helloRecord)
	if err != nil {
		return batchSize, fmt.Errorf("failed to parse hello: %w", err)
	}
	batchSize = hello.BatchSize

	bets := make([]lottery.Bet, 0, batchSize)

	for i := 0; i < hello.BatchSize; i++ {
		var localBetRecord []string
		localBetRecord, err = reader.Read()
		if err != nil {
			return
		}
		localBet, err := common.LocalBetFromRecord(localBetRecord)
		if err != nil {
			return batchSize, fmt.Errorf("failed to parse bet: %w", err)
		}
		bet := lottery.Bet{
			Agency:   hello.AgencyId,
			LocalBet: localBet,
		}

		bets = append(bets, bet)
	}

	writer := csv.NewWriter(conn)

	err = lottery.StoreBets(bets)
	if err != nil {
		storeErr := fmt.Errorf("failed to store bets: %w", err)

		_ = writer.Write(common.Err{}.ToRecord())
		writer.Flush()
		sendErr := writer.Error()
		if err != nil {
			sendErr = fmt.Errorf("failed to send err message: %w", err)
		}

		return batchSize, errors.Join(storeErr, sendErr)
	}

	_ = writer.Write(common.Ok{}.ToRecord())
	writer.Flush()
	err = writer.Error()
	if err != nil {
		return batchSize, fmt.Errorf("failed to send ok: %w", err)
	}

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
