package log

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func SetupLogger(loglevel string, format string) error {
	loglevel = strings.ToLower(loglevel)

	var level slog.Level

	switch loglevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "info":
		level = slog.LevelInfo
	default:
		return fmt.Errorf("unknown log level '%s'", loglevel)
	}

	var handler slog.Handler
	switch format {
	case "console":
		noColor := determineNoColor()
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		handler = &consoleHandler{level: level, out: os.Stdout, errout: os.Stderr, noColor: noColor, cwd: cwd}
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	default:
		return fmt.Errorf("unknown log formatter '%s'", format)
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return nil
}
