package main

import (
	"strings"
	"time"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
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
}

func initConfig() (config, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigFile("./config.yaml")
	_ = v.ReadInConfig()

	var c config
	err := v.Unmarshal(&c)

	return c, err
}

func logConfig(c config) {
	log.Infof(common.FmtLog("action", "config",
		"result", "success",
		"server.address", c.Server.Address,
		"loop.amount", c.Loop.Amount,
		"loop.period", c.Loop.Period,
		"log.level", c.Log.Level,
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
	client.startClientLoop()
}
