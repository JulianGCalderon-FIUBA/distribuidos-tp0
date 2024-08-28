package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/op/go-logging"
)

func InitLogger(logLevel string) error {
	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}

	base := logging.NewLogBackend(os.Stdout, "", 0)

	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:-8s} %{message}`,
	)
	formatter := logging.NewBackendFormatter(base, format)

	leveled := logging.AddModuleLevel(formatter)
	leveled.SetLevel(logLevelCode, "")

	logging.SetBackend(leveled)

	return nil
}

func FmtLog(data ...any) string {
	listed := make([]string, 0, len(data))

	for len(data) >= 2 {
		listed = append(listed, fmt.Sprintf("%v: %v", data[0], data[1]))
		data = data[2:]
	}

	return strings.Join(listed, " | ")
}
