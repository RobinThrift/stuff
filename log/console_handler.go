package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"log/slog"
)

type consoleHandler struct {
	level  slog.Level
	out    io.Writer
	errout io.Writer

	noColor bool

	group string
	attrs []slog.Attr
}

func (h *consoleHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *consoleHandler) Handle(_ context.Context, record slog.Record) error {
	out := h.out

	if record.Level >= slog.LevelError {
		out = h.errout
	}

	var msg bytes.Buffer

	color := ""
	if record.Level == slog.LevelWarn && !h.noColor {
		color = "\x1b[33m" // yellow
	}

	if record.Level >= slog.LevelError && !h.noColor {
		color = "\x1b[31m" // red
	}

	if record.Level == slog.LevelDebug && !h.noColor {
		color = "\x1b[34m" // blue
	}

	msg.WriteString(color)

	if h.level <= slog.LevelDebug {
		msg.WriteString(record.Time.Format(time.TimeOnly))
		msg.WriteString(" ")
	}

	if h.level <= slog.LevelDebug && record.PC != 0 {
		f := runtime.FuncForPC(record.PC)
		file, line := f.FileLine(record.PC)

		file = strings.Replace(file, "github.com/kodeshack/stuff/", "", 1)

		msg.WriteString(file)
		msg.WriteString(":")
		msg.WriteString(fmt.Sprint(line))
		msg.WriteString(" ")
	}

	msg.WriteString(record.Message)

	msg.WriteString("\n")

	if color != "" && !h.noColor {
		_, err := msg.WriteString("\x1b[0m") // reset
		if err != nil {
			return err
		}
	}

	_, err := out.Write(msg.Bytes())
	return err
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nextAttrs := h.attrs
	if h.group != "" {
		args := make([]any, 0, len(attrs))
		for _, attr := range attrs {
			args = append(args, attr)
		}
		nextAttrs = append(nextAttrs, slog.Group(h.group, args...))
	} else {
		nextAttrs = append(nextAttrs, attrs...)
	}

	return &consoleHandler{
		level:  h.level,
		out:    h.out,
		errout: h.errout,
		attrs:  nextAttrs,
	}
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	return &consoleHandler{
		level:  h.level,
		out:    h.out,
		errout: h.errout,
		group:  name,
		attrs:  h.attrs,
	}
}

func determineNoColor() bool {
	if asBool, err := strconv.ParseBool(os.Getenv("CLICOLOR")); err == nil {
		return !asBool
	}

	if asBool, err := strconv.ParseBool(os.Getenv("SF_CLI_NO_COLOR")); err == nil {
		return !asBool
	}

	return false
}
