package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Init creates a *slog.Logger based on level and format.
// level: debug, info, warn, error, off
// format: pretty (default), json
func Init(level, format string) *slog.Logger {
	var minLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		minLevel = slog.LevelDebug
	case "warn", "warning":
		minLevel = slog.LevelWarn
	case "error":
		minLevel = slog.LevelError
	case "off":
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	default:
		minLevel = slog.LevelInfo
	}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: minLevel})
	default:
		handler = newPrettyHandler(os.Stderr, minLevel)
	}

	return slog.New(handler)
}

// ColorizeStatus returns a colorized HTTP status code string for terminal output.
// 2xx = green, 4xx = yellow, 5xx = red.
func ColorizeStatus(status int) string {
	var color string
	switch {
	case status >= 200 && status < 300:
		color = "\033[32m"
	case status >= 400 && status < 500:
		color = "\033[33m"
	case status >= 500:
		color = "\033[31m"
	default:
		color = "\033[0m"
	}
	return fmt.Sprintf("%s%d\033[0m", color, status)
}

type prettyHandler struct {
	w     io.Writer
	mu    sync.Mutex
	level slog.Level
	attrs []slog.Attr
}

func newPrettyHandler(w io.Writer, level slog.Level) *prettyHandler {
	return &prettyHandler{w: w, level: level}
}

func (h *prettyHandler) clone() *prettyHandler {
	return &prettyHandler{
		w:     h.w,
		level: h.level,
		attrs: append([]slog.Attr(nil), h.attrs...),
	}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var buf bytes.Buffer

	label, color := levelLabel(r.Level)
	buf.WriteString(color)
	buf.WriteString(label)
	buf.WriteString("\033[0m")
	buf.WriteByte(' ')

	buf.WriteString(r.Message)

	for _, a := range h.attrs {
		buf.WriteByte(' ')
		writeAttr(&buf, a)
	}

	r.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(' ')
		writeAttr(&buf, a)
		return true
	})

	buf.WriteByte('\n')
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := h.clone()
	c.attrs = append(c.attrs, attrs...)
	return c
}

func (h *prettyHandler) WithGroup(_ string) slog.Handler {
	return h.clone()
}

func levelLabel(l slog.Level) (string, string) {
	switch {
	case l < slog.LevelInfo:
		return "DBG", "\033[90m"
	case l == slog.LevelInfo:
		return "INF", "\033[32m"
	case l == slog.LevelWarn:
		return "WRN", "\033[33m"
	default:
		return "ERR", "\033[31m"
	}
}

func writeAttr(buf *bytes.Buffer, a slog.Attr) {
	buf.WriteString("\033[36m")
	buf.WriteString(a.Key)
	buf.WriteString("\033[0m=")
	writeValue(buf, a.Value)
}

func writeValue(buf *bytes.Buffer, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		s := v.String()
		if needsQuote(s) {
			buf.WriteByte('"')
			buf.WriteString(s)
			buf.WriteByte('"')
		} else {
			buf.WriteString(s)
		}
	case slog.KindInt64:
		buf.WriteString(strconv.FormatInt(v.Int64(), 10))
	case slog.KindUint64:
		buf.WriteString(strconv.FormatUint(v.Uint64(), 10))
	case slog.KindFloat64:
		buf.WriteString(strconv.FormatFloat(v.Float64(), 'f', -1, 64))
	case slog.KindBool:
		buf.WriteString(strconv.FormatBool(v.Bool()))
	case slog.KindDuration:
		buf.WriteString(v.Duration().String())
	case slog.KindTime:
		buf.WriteString(v.Time().Format("15:04:05"))
	case slog.KindAny:
		a := v.Any()
		if err, ok := a.(error); ok {
			s := err.Error()
			if needsQuote(s) {
				buf.WriteByte('"')
				buf.WriteString(s)
				buf.WriteByte('"')
			} else {
				buf.WriteString(s)
			}
		} else {
			s := fmt.Sprint(a)
			if needsQuote(s) {
				buf.WriteByte('"')
				buf.WriteString(s)
				buf.WriteByte('"')
			} else {
				buf.WriteString(s)
			}
		}
	case slog.KindGroup:
		// pretty handler ignores groups
	}
}

func needsQuote(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' || s[i] == '"' {
			return true
		}
	}
	return false
}
