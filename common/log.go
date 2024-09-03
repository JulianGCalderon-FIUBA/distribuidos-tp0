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

func FmtLog(action string, err error, data ...any) string {
	listed := make([]string, 0)

	listed = append(listed, fmt.Sprintf("action: %v", action))
	switch err {
	case nil:
		listed = append(listed, "result: success")
	default:
		listed = append(listed, "result: fail")
		listed = append(listed, fmt.Sprintf("error: %v", err))
	}

	for len(data) >= 2 {
		listed = append(listed, fmt.Sprintf("%v: %v", data[0], data[1]))
		data = data[2:]
	}

	return strings.Join(listed, " | ")
}
