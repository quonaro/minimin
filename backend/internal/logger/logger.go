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
// format: json, text (default text)
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
		handler = newCompactHandler(os.Stderr, minLevel, true)
	}

	return slog.New(handler)
}

type compactHandler struct {
	w      io.Writer
	mu     sync.Mutex
	level  slog.Level
	colors bool
	attrs  []slog.Attr
}

func newCompactHandler(w io.Writer, level slog.Level, colors bool) *compactHandler {
	return &compactHandler{w: w, level: level, colors: colors}
}

func (h *compactHandler) clone() *compactHandler {
	return &compactHandler{
		w:      h.w,
		level:  h.level,
		colors: h.colors,
		attrs:  append([]slog.Attr(nil), h.attrs...),
	}
}

func (h *compactHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *compactHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var buf bytes.Buffer

	buf.WriteString(r.Time.Format("15:04:05"))
	buf.WriteByte(' ')

	lvl, color := levelLabel(r.Level)
	if h.colors {
		buf.WriteString(color)
	}
	buf.WriteString(lvl)
	if h.colors {
		buf.WriteString("\033[0m")
	}
	buf.WriteString(" | ")
	buf.WriteString(r.Message)

	for _, a := range h.attrs {
		buf.WriteByte(' ')
		writeValue(&buf, a.Value)
	}

	r.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(' ')
		writeValue(&buf, a.Value)
		return true
	})

	buf.WriteByte('\n')
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *compactHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := h.clone()
	c.attrs = append(c.attrs, attrs...)
	return c
}

func (h *compactHandler) WithGroup(_ string) slog.Handler {
	return h.clone()
}

func levelLabel(l slog.Level) (string, string) {
	switch {
	case l < slog.LevelInfo:
		return "D", "\033[90m"
	case l == slog.LevelInfo:
		return "I", "\033[32m"
	case l == slog.LevelWarn:
		return "W", "\033[33m"
	default:
		return "E", "\033[31m"
	}
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
		// compact handler ignores groups
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
