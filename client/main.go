package main

import (
	"context"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
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
	Loop struct {
		Amount int
		Period time.Duration
	}
	Log struct {
		Level string
	}
	Batch struct {
		MaxAmount int
	}
	Bet struct {
		FirstName string
		LastName  string
		Document  int
		Birthdate time.Time
		Number    int
	}
}

func initConfig() (config, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = v.BindEnv("bet.firstName", "NOMBRE")
	_ = v.BindEnv("bet.lastName", "APELLIDO")
	_ = v.BindEnv("bet.document", "DOCUMENTO")
	_ = v.BindEnv("bet.birthdate", "NACIMIENTO")
	_ = v.BindEnv("bet.number", "NUMERO")

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
	log.Infof(common.FmtLog("action", "config",
		"result", "success",
		"server.address", c.Server.Address,
		"loop.amount", c.Loop.Amount,
		"loop.period", c.Loop.Period,
		"log.level", c.Log.Level,
		"bet.firstName", c.Bet.FirstName,
		"bet.lastName", c.Bet.LastName,
		"bet.document", c.Bet.Document,
		"bet.birthdate", c.Bet.Birthdate,
		"bet.number", c.Bet.Number,
	))
}

func main() {
	c, err := initConfig()
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = common.InitLogger(c.Log.Level)
	if err != nil {
		log.Fatalf("%s", err)
	}

	logConfig(c)

	clientConfig := clientConfig{
		serverAddress: c.Server.Address,
		id:            c.Id,
		loopAmount:    c.Loop.Amount,
		loopPeriod:    c.Loop.Period,
	}
	client := newClient(clientConfig)

	ctx, cancel_handler := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel_handler()

	client.startClientLoop(ctx)
}
