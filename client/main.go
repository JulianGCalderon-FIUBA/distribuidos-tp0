package main

import (
	"strings"
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
	Log struct {
		Level string
	}
	Batch struct {
		MaxAmount int
	}
	Bet common.LocalBet
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
	}
	client := newClient(clientConfig)

	err = client.sendBet(c.Bet)
	if err != nil {
		log.Error(common.FmtLog("action", "apuesta_enviada",
			"result", "fail",
			"error", err,
		))
	} else {
		log.Info(common.FmtLog("action", "apuesta_enviada",
			"result", "success",
			"dni", c.Bet.Document,
			"numero", c.Bet.Number,
		))
	}
}
