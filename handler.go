package hue

import (
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type hueHandler struct {
	writer io.Writer
	mx     sync.Mutex

	level       slog.Level
	timeFormat  string
	replaceAttr func(groups []string, a slog.Attr) slog.Attr

	group string
}

// NewHueHandler creates a [slog.Handler] that writes pretty formatted logs to the given writer.
func NewHueHandler(writer io.Writer, options *Options) slog.Handler {
	h := &hueHandler{
		writer:     writer,
		timeFormat: DefaultTimeFormat,
		level:      DefaultLogLevel,
	}

	if options == nil {
		return h
	}

	h.replaceAttr = options.ReplaceAttr

	if options.TimeFormat != "" {
		h.timeFormat = options.TimeFormat
	}

	if options.Level != nil {
		h.level = options.Level.Level()
	}

	return h
}

// Enabled implements slog.Handler.
func (h *hueHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements slog.Handler.
func (h *hueHandler) Handle(ctx context.Context, rec slog.Record) error {
	buf := &buffer{}

	// write the time
	if !rec.Time.IsZero() {
		rec.Time.Round(0)
		h.writeTime(buf, rec.Time)
		buf.WriteString(" ")
	}

	// write the level
	h.writeLevel(buf, rec.Level)
	buf.WriteString(" ")

	// write the message
	buf.WriteString(rec.Message)
	buf.WriteString(" ")

	rec.Attrs(func(a slog.Attr) bool {
		h.writeAttr(buf, a, h.group)
		return true
	})

	buf.WriteString("\n")

	h.mx.Lock()
	defer h.mx.Unlock()

	_, err := h.writer.Write(*buf)
	return err
}

// WithAttrs implements slog.Handler.
func (h *hueHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("unimplemented")
}

// WithGroup implements slog.Handler.
func (h *hueHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &hueHandler{
		writer:      h.writer,
		level:       h.level,
		timeFormat:  h.timeFormat,
		replaceAttr: h.replaceAttr,
		group:       h.group + name + ".",
	}
}

func (h *hueHandler) writeTime(buf *buffer, t time.Time) {
	if h.replaceAttr != nil {
		attr := h.replaceAttr(nil, slog.Time(slog.TimeKey, t))
		if attr.Value.Kind() == slog.KindTime {
			t = attr.Value.Time()
		} else {
			h.writeAttr(buf, attr, "")
			return
		}
	}

	buf.WriteString(timeStyle.Render(t.Format(h.timeFormat)))
}

func (h *hueHandler) writeLevel(buf *buffer, level slog.Level) {
	if h.replaceAttr != nil {
		attr := h.replaceAttr(nil, slog.Any(slog.LevelKey, level))
		buf.WriteString(attr.Value.String())
		return
	}

	switch level {
	case slog.LevelDebug:
		buf.WriteString(debugLevelStyle.Render("DBG"))
	case slog.LevelInfo:
		buf.WriteString(infoLevelStyle.Render("INF"))
	case slog.LevelWarn:
		buf.WriteString(warnLevelStyle.Render("WRN"))
	case slog.LevelError:
		buf.WriteString(errorLevelStyle.Render("ERR"))
	}
}

func (h *hueHandler) writeAttr(buf *buffer, attr slog.Attr, prefix string) {
	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			prefix = prefix + attr.Key + "."
		}

		for _, groupAttrs := range attr.Value.Group() {
			h.writeAttr(buf, groupAttrs, prefix)
		}

		return
	}

	h.writeAttrKey(buf, attr, prefix)
	h.writeAttrValue(buf, attr)
	buf.WriteString(" ")
}

func (h *hueHandler) writeAttrKey(buf *buffer, attr slog.Attr, prefix string) {
	style := attrStyle
	if styledVal, ok := attr.Value.Any().(StyledValue); ok {
		style = styledVal.Style()
	}

	buf.WriteString(fmt.Sprintf("%s=", style.Render(prefix+attr.Key)))
}

func (h *hueHandler) writeAttrValue(buf *buffer, attr slog.Attr) {
	switch attr.Value.Kind() {
	case slog.KindString:
		*buf = strconv.AppendQuote(*buf, attr.Value.String())
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, attr.Value.Bool())
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, attr.Value.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, attr.Value.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, attr.Value.Float64(), 'f', -1, 64)
	case slog.KindTime:
		*buf = strconv.AppendQuote(*buf, attr.Value.Time().String())
	case slog.KindDuration:
		*buf = append(*buf, attr.Value.Duration().String()...)
	case slog.KindAny:
		switch avt := attr.Value.Any().(type) {
		case encoding.TextMarshaler:
			enc, err := avt.MarshalText()
			if err != nil {
				break
			}
			*buf = strconv.AppendQuote(*buf, string(enc))
		case fmt.Stringer:
			*buf = strconv.AppendQuote(*buf, avt.String())
		default:
			*buf = strconv.AppendQuote(*buf, fmt.Sprintf("%+v", avt))
		}
	}
}
