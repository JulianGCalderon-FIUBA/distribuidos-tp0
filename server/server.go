package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/server/lottery"
)

const MAX_AGENCIES = 5

type server struct {
	listener       net.Listener
	storageLock    *sync.Mutex
	lotteryFinish  *sync.WaitGroup
	activeHandlers *sync.WaitGroup
}

func newServer(port int, listenBacklog int) (*server, error) {
	_ = listenBacklog
	address := fmt.Sprintf("0.0.0.0:%v", port)
	listener, err := net.Listen("tcp", address)
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
	listenerCloser := common.SpawnCloser(ctx, s.listener, closeListener)
	defer func() {
		closeErr := listenerCloser.Close()
		err = errors.Join(err, closeErr)
	}()

	handlerCtx, cancelHandlerCtx := context.WithCancel(ctx)
	defer func() {
		cancelHandlerCtx()
		s.activeHandlers.Wait()
	}()

	for {
		h, err := s.acceptClient()
		if err != nil {
			return err
		}
		if h != nil {
			s.activeHandlers.Add(1)
			go func(h *handler) {
				err := h.run(handlerCtx)
				if err != nil {
					if !errors.Is(err, net.ErrClosed) {
						log.Error(common.FmtLog("handle_client", err))
					}
				}
				s.activeHandlers.Done()
			}(h)
		}
	}
}

func (s *server) acceptClient() (*handler, error) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return nil, err
		}

		h, err := createHandler(s, conn)
		if err != nil {
			log.Error(common.FmtLog("handshake", err))
			continue
		}

		log.Info(common.FmtLog("handshake", nil,
			"agency_id", h.agencyId,
		))

		return h, nil
	}
}

func closeListener(listener net.Listener) error {
	err := listener.Close()
	if err != nil {
		log.Error(common.FmtLog("close_listener", err))
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
