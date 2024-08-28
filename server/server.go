package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
)

type server struct {
	listener net.Listener
}

func newServer(port int, listenBacklog int) (*server, error) {
	_ = listenBacklog
	address := fmt.Sprintf("0.0.0.0:%v", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &server{
		listener: listener,
	}, nil
}

func (s *server) run() error {
	for {
		log.Info(common.FmtLog("action", "accept_connections",
			"result", "in_progress",
		))

		conn, err := s.listener.Accept()
		if err != nil {
			return err
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

func (s *server) handleClientConnection(conn net.Conn) error {
	defer conn.Close()

	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}

	addr, err := net.ResolveTCPAddr(conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	if err != nil {
		return err
	}

	log.Info(common.FmtLog("action", "receive_message",
		"result", "success",
		"ip", addr.IP,
		"msg", msg,
	))

	writer := bufio.NewWriter(conn)
	_, _ = writer.WriteString(msg)
	err = writer.Flush()

	return err
}
