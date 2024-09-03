package main

import (
	"context"
	"errors"
	"net"
	"os/signal"
	"syscall"

	"github.com/juliangcalderon-fiuba/distribuidos-tp0/common"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

var log = logging.MustGetLogger("log")

type config struct {
	Default struct {
		Server_Port           int
		Server_Ip             string
		Server_Listen_Backlog int
		Logging_Level         string
	}
}

func initConfig() (config, error) {
	v := viper.New()

	_ = v.BindEnv("default.server_port", "SERVER_PORT")
	_ = v.BindEnv("default.server_ip", "SERVER_IP")
	_ = v.BindEnv("default.server_listen_backlog", "SERVER_LISTEN_BACKLOG")
	_ = v.BindEnv("default.logging_level", "LOGGING_LEVEL")

	v.SetConfigFile("./config.ini")
	_ = v.ReadInConfig()

	var c config
	err := v.Unmarshal(&c)

	return c, err
}

func logConfig(c config) {
	log.Infof(common.FmtLog("action", "config",
		"result", "success",
		"server.ip", c.Default.Server_Ip,
		"server.port", c.Default.Server_Port,
		"server.listen_backlog", c.Default.Server_Listen_Backlog,
		"logging.level", c.Default.Logging_Level,
	))
}

func main() {
	c, err := initConfig()
	if err != nil {
		log.Fatalf("failed to init config: %s", err)
	}

	err = common.InitLogger(c.Default.Logging_Level)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	logConfig(c)

	s, err := newServer(c.Default.Server_Port, c.Default.Server_Listen_Backlog)
	if err != nil {
		log.Fatalf("failed to create server: %s", err)
	}

	ctx, cancel_handler := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel_handler()

	err = s.run(ctx)
	if err != nil {
		if !errors.Is(err, net.ErrClosed) {
			log.Fatalf("failed to run server: %s", err)
		}
	}
}
