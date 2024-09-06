package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/juliangcalderon-fiuba/distribuidos-tp0/safeio"
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
	Loop struct {
		Period time.Duration
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
		"loop.period", c.Loop.Period,
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

	betsPath := fmt.Sprintf(".data/agency-%v.csv", c.Id)
	betsFile, err := os.Open(betsPath)
	if err != nil {
		log.Fatalf("Failed to open bet dataset: %v", err)
	}
	betsReader := safeio.NewReader(betsFile)

	clientConfig := clientConfig{
		serverAddress: c.Server.Address,
		batchSize:     c.Batch.MaxAmount,
		id:            c.Id,
		loopPeriod:    c.Loop.Period,
	}
	client := newClient(clientConfig, betsReader)

	ctx, ctx_cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer ctx_cancel()

	err = client.run(ctx)
	if err != nil {
		if !errors.Is(err, net.ErrClosed) {
			log.Fatalf("failed to run client: %s", err)
		}
	}
}
