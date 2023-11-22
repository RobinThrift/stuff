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

	"github.com/RobinThrift/stuff/internal/requestid"
)

type consoleHandler struct {
	cwd string

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

func (h *consoleHandler) Handle(ctx context.Context, record slog.Record) error {
	out := h.out

	if reqID, ok := requestid.FromCtx(ctx); ok {
		record.AddAttrs(slog.String("request_id", reqID))
	}

	if record.Level >= slog.LevelError {
		out = h.errout
	}

	var msg bytes.Buffer

	if h.level <= slog.LevelDebug {
		msg.WriteString(record.Time.Format(time.TimeOnly))
		msg.WriteString(" ")
	}

	if !h.noColor {
		color := ""
		switch record.Level {
		case slog.LevelDebug:
			color = "\x1b[34m" // blue
		case slog.LevelInfo:
			color = "\x1b[32m" // green
		case slog.LevelWarn:
			color = "\x1b[33m" // yellow
		case slog.LevelError:
			color = "\x1b[31m" // red
		}

		msg.WriteString(color)
		msg.WriteString(record.Level.String())
		msg.WriteString("\x1b[0m ") // reset
	}

	if h.level <= slog.LevelDebug && record.PC != 0 {
		f := runtime.FuncForPC(record.PC)
		file, line := f.FileLine(record.PC)

		file = strings.Replace(file, h.cwd+"/", "", 1)

		if !h.noColor {
			msg.WriteString("\x1b[90m") // reset
		}

		msg.WriteString(file)
		msg.WriteString(":")
		msg.WriteString(fmt.Sprint(line))
		msg.WriteString(" ")

		if !h.noColor {
			msg.WriteString("\x1b[0m") // reset
		}
	}

	msg.WriteString(record.Message)

	record.Attrs(func(a slog.Attr) bool {
		msg.WriteRune(' ')
		if a.Key == "error" && !h.noColor {
			msg.WriteString("\x1b[31m")
			msg.WriteString(a.Key)
			msg.WriteString("\x1b[0m")
		} else {
			msg.WriteString(a.Key)
		}

		msg.WriteString(`="`)
		msg.WriteString(a.Value.String())
		msg.WriteRune('"')

		return true
	})

	msg.WriteString("\n")

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
	if asBool, err := strconv.ParseBool(os.Getenv("NO_COLOR")); err == nil {
		return asBool
	}

	return false
}
