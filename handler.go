package hue

import (
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	DefaultTimeFormat = time.TimeOnly
	DefaultLogLevel   = slog.LevelInfo
)

type hueHandler struct {
	writer io.Writer
	mx     sync.Mutex

	level       slog.Level
	timeFormat  string
	replaceAttr func(groups []string, a slog.Attr) slog.Attr
	withCaller  bool
	withPrefix  bool

	group string
	attrs []slog.Attr
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
	h.withCaller = options.WithCaller
	h.withPrefix = options.WithPrefix

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

	rec.AddAttrs(h.attrs...)

	// write the time
	if !rec.Time.IsZero() {
		rec.Time.Round(0)
		h.writeTime(buf, rec.Time)
		buf.WriteString(" ")
	}

	// write the level
	h.writeLevel(buf, rec.Level)
	buf.WriteString(" ")

	if h.withCaller {
		h.writeCaller(buf, rec.PC)
		buf.WriteString(" ")
	}

	if h.withPrefix {
		rec = h.writePrefix(buf, rec)
	}

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
	if len(attrs) == 0 {
		return h
	}

	return &hueHandler{
		writer:      h.writer,
		level:       h.level,
		timeFormat:  h.timeFormat,
		replaceAttr: h.replaceAttr,
		withCaller:  h.withCaller,
		withPrefix:  h.withPrefix,
		group:       h.group,
		attrs:       append(h.attrs, attrs...),
	}
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
		withCaller:  h.withCaller,
		withPrefix:  h.withPrefix,
		group:       h.group + name + ".",
		attrs:       h.attrs,
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

	buf.WriteString(mutedStyle.Render(t.Format(h.timeFormat)))
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
	if styledVal, ok := attr.Value.Any().(StyledAttr); ok {
		style = styledVal.Style().Faint(true)
	}

	buf.WriteString(style.Render(fmt.Sprintf("%s=", prefix+attr.Key)))
}

func (h *hueHandler) writeAttrValue(buf *buffer, attr slog.Attr) {
	style := h.attrStyle(attr, lipgloss.NewStyle())
	h.writeStyledAttrValue(buf, attr, style, true)
}

func (h *hueHandler) attrStyle(attr slog.Attr, defaultStyle lipgloss.Style) lipgloss.Style {
	if styledVal, ok := attr.Value.Any().(StyledAttr); ok {
		return styledVal.Style()
	}

	return defaultStyle
}

func (h *hueHandler) writeStyledAttrValue(buf *buffer, attr slog.Attr, style lipgloss.Style, quote bool) {
	formatter := func(value string) string {
		if quote {
			return strconv.Quote(value)
		}
		return value
	}

	switch attr.Value.Kind() {
	case slog.KindString:
		*buf = append(*buf, style.Render(formatter(attr.Value.String()))...)
	case slog.KindBool:
		*buf = append(*buf, style.Render(strconv.FormatBool(attr.Value.Bool()))...)
	case slog.KindInt64:
		*buf = append(*buf, style.Render(strconv.FormatInt(attr.Value.Int64(), 10))...)
	case slog.KindUint64:
		*buf = append(*buf, style.Render(strconv.FormatUint(attr.Value.Uint64(), 10))...)
	case slog.KindFloat64:
		*buf = append(*buf, style.Render(strconv.FormatFloat(attr.Value.Float64(), 'f', -1, 64))...)
	case slog.KindTime:
		*buf = append(*buf, style.Render(formatter(attr.Value.Time().String()))...)
	case slog.KindDuration:
		*buf = append(*buf, style.Render(attr.Value.Duration().String())...)
	case slog.KindAny:
		switch avt := attr.Value.Any().(type) {
		case encoding.TextMarshaler:
			enc, err := avt.MarshalText()
			if err != nil {
				break
			}
			*buf = append(*buf, style.Render(formatter(string(enc)))...)
		case fmt.Stringer:
			*buf = append(*buf, style.Render(formatter(avt.String()))...)
		default:
			*buf = append(*buf, style.Render(formatter(fmt.Sprintf("%+v", avt)))...)
		}
	}
}

func (h *hueHandler) writeCaller(buf *buffer, pc uintptr) {
	// grab the caller from the stack
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()

	src := slog.Source{
		Function: frame.Function,
		File:     frame.File,
		Line:     frame.Line,
	}

	if h.replaceAttr != nil {
		attr := h.replaceAttr(nil, slog.Any("caller", &src))
		if v, ok := attr.Value.Any().(*slog.Source); ok {
			src = *v
		}
	}

	_, file := filepath.Split(src.File)

	// write the caller
	buf.WriteString(mutedStyle.Render(fmt.Sprintf("<%s:%d>", file, src.Line)))
}

func (h *hueHandler) writePrefix(buf *buffer, rec slog.Record) slog.Record {
	// early return if theres no attributes on the record
	if rec.NumAttrs() == 0 {
		return rec
	}

	// find the last attribute marked as a prefix
	var prefix *slog.Attr
	attrs := make([]slog.Attr, 0, rec.NumAttrs())
	rec.Attrs(func(a slog.Attr) bool {
		if _, ok := a.Value.Any().(PrefixAttr); ok && prefix == nil {
			prefix = &a
		} else {
			attrs = append(attrs, a)
		}

		return true
	})

	if prefix == nil {
		return rec
	}

	style := h.attrStyle(*prefix, mutedStyle)
	h.writeStyledAttrValue(buf, *prefix, style, false)
	buf.WriteString(style.Render(":"))
	buf.WriteString(" ")

	// returned a cloned record without the attribute that was used as a prefix
	newRec := slog.Record{Time: rec.Time, Level: rec.Level, Message: rec.Message}
	newRec.AddAttrs(attrs...)
	return newRec
}
