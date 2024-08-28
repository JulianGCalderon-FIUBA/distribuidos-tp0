package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
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

		err = s.handleClientConnection(conn)
		if err != nil {
			log.Error(common.FmtLog("action", "receive_message",
				"result", "fail",
				"error", err,
			))
		}
	}

}

func (s *server) handleClientConnection(conn net.Conn) (err error) {
	defer func() {
		closeErr := closeConnection(conn)
		err = errors.Join(err, closeErr)
	}()

	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return
	}

	addr, err := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	if err != nil {
		return
	}

	log.Info(common.FmtLog("action", "receive_message",
		"result", "success",
		"ip", addr.IP,
		"msg", msg,
	))

	writer := bufio.NewWriter(conn)
	_, _ = writer.WriteString(msg)
	err = writer.Flush()

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

	log.Info(common.FmtLog("action", "close_connection",
		"result", "success",
	))
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

	log.Info(common.FmtLog("action", "close_listener",
		"result", "success",
	))
	return nil
}
