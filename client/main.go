package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/protocol"
	"github.com/mitchellh/mapstructure"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

var log = logging.MustGetLogger("log")

type config struct {
	Id     int
	Server struct {
		Address string
	}
	Log struct {
		Level string
	}
	Batch struct {
		MaxAmount int
	}
}

func initConfig() (config, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigFile("./config.yaml")
	_ = v.ReadInConfig()

	var c config
	err := v.Unmarshal(&c, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.DateOnly),
			mapstructure.StringToTimeDurationHookFunc(),
		)))

	return c, err
}

func logConfig(c config) {
	log.Infof(common.FmtLog("config", nil,
		"server.address", c.Server.Address,
		"batch.maxAmount", c.Batch.MaxAmount,
		"log.level", c.Log.Level,
	))
}

func main() {
	c, err := initConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	err = common.InitLogger(c.Log.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logConfig(c)

	bets, err := readAgency(c.Id)
	if err != nil {
		log.Fatalf("Failed to read bet dataset: %v", err)
	}

	clientConfig := clientConfig{
		serverAddress: c.Server.Address,
		batchSize:     c.Batch.MaxAmount,
		id:            c.Id,
	}
	client := newClient(clientConfig)

	ctx, ctx_cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer ctx_cancel()

	err = client.run(ctx, bets)
	if err != nil {
		if !errors.Is(err, net.ErrClosed) {
			log.Fatalf("failed to run client: %s", err)
		}
	}
}

func readAgency(id int) (bets []protocol.BetMessage, err error) {
	agencyPath := fmt.Sprintf("./.data/agency-%v.csv", id)
	file, err := os.Open(agencyPath)
	if err != nil {
		return
	}
	defer func() {
		closeErr := file.Close()
		err = errors.Join(err, closeErr)
	}()

	reader := csv.NewReader(file)
	bets = make([]protocol.BetMessage, 0)

	for {
		var row []string
		row, err = reader.Read()
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
		if err != nil {
			return
		}

		var bet protocol.BetMessage
		bet, err = protocol.Deserialize[protocol.BetMessage](row)
		if err != nil {
			return
		}

		bets = append(bets, bet)
	}

	return
}
