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
				log.Infof(common.FmtLog("action", "shutdown"))
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

		bet, err := s.receiveBet(conn)
		if err != nil {
			log.Error(common.FmtLog("action", "apuesta_almacenada",
				"result", "fail",
				"error", err,
			))
		} else {
			log.Info(common.FmtLog("action", "apuesta_almacenada",
				"result", "success",
				"dni", bet.Document,
				"numero", bet.Number,
			))
		}
	}
}

func (s *server) receiveBet(conn net.Conn) (bet lottery.Bet, err error) {
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
		return bet, fmt.Errorf("failed to parse hello: %w", err)
	}

	localBetRecord, err := reader.Read()
	if err != nil {
		return
	}
	localBet, err := common.LocalBetFromRecord(localBetRecord)
	if err != nil {
		return bet, fmt.Errorf("failed to parse bet: %w", err)
	}

	bet = lottery.Bet{
		Agency:   hello.AgencyId,
		LocalBet: localBet,
	}

	err = lottery.StoreBets([]lottery.Bet{bet})
	if err != nil {
		return bet, fmt.Errorf("failed to store bet: %w", err)
	}

	writer := csv.NewWriter(conn)
	_ = writer.Write([]string{"OK"})
	writer.Flush()
	err = writer.Error()
	if err != nil {
		return bet, fmt.Errorf("failed to send ok: %w", err)
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
