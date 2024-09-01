package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/server/lottery"
)

const MAX_AGENCIES = 5

type server struct {
	listener       *net.TCPListener
	storageLock    *sync.Mutex
	lotteryFinish  *sync.WaitGroup
	activeHandlers *sync.WaitGroup
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

	lotteryFinish := &sync.WaitGroup{}
	lotteryFinish.Add(MAX_AGENCIES)

	return &server{
		listener:       listener,
		lotteryFinish:  lotteryFinish,
		storageLock:    &sync.Mutex{},
		activeHandlers: &sync.WaitGroup{},
	}, nil
}

func (s *server) run(ctx context.Context) (err error) {
	defer func() {
		closeErr := closeListener(s.listener)
		err = errors.Join(err, closeErr)
	}()

	handlerCtx, cancelHandlerCtx := context.WithCancel(ctx)
	defer func() {
		s.activeHandlers.Wait()
		cancelHandlerCtx()
	}()

	for i := 0; i < MAX_AGENCIES; i++ {
		h, err := s.acceptClient(ctx)
		if err != nil {
			log.Critical(common.FmtLog("action", "accept",
				"result", "fail",
				"error", err,
			))
			cancelHandlerCtx()
			return err
		}

		s.activeHandlers.Add(1)
		go func(h *handler) {
			_ = h.run(handlerCtx)
			s.activeHandlers.Done()
		}(h)
	}

	return nil
}

func (s *server) acceptClient(ctx context.Context) (*handler, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}

		err := s.listener.SetDeadline(time.Now().Add(500 * time.Millisecond))
		if err != nil {
			return nil, err
		}

		conn, err := s.listener.AcceptTCP()
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		}
		if err != nil {
			return nil, err
		}

		h, err := createHandler(s, conn)
		if err != nil {
			log.Error(common.FmtLog("action", "handshake",
				"result", "fail",
				"error", err,
			))
			continue
		}

		log.Error(common.FmtLog("action", "handshake",
			"result", "success",
			"agency_id", h.agencyId,
		))

		return h, nil
	}
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

func (s *server) getWinners() (map[int][]int, error) {
	s.storageLock.Lock()
	allBets, err := lottery.LoadBets()
	s.storageLock.Unlock()

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
