package hue

import (
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const DefaultLogLevel = slog.LevelInfo
const DefaultTimeFormat = time.TimeOnly

type hueHandler struct {
	w  io.Writer
	mx *sync.Mutex

	opts Options

	group  string
	prefix buffer
	attrs  buffer
}

func New(w io.Writer, options *Options) *hueHandler {
	h := &hueHandler{
		w:  w,
		mx: &sync.Mutex{},
		opts: Options{
			Level:      DefaultLogLevel,
			TimeFormat: DefaultTimeFormat,
			WithPrefix: true,
			WithCaller: false,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				return a
			},
		},
	}

	if options != nil {
		h.opts = *options
	}

	return h
}

func (h *hueHandler) clone() *hueHandler {
	return &hueHandler{
		w:  h.w,
		mx: h.mx,

		opts: h.opts,

		group:  h.group,
		prefix: slices.Clip(h.prefix),
		attrs:  slices.Clip(h.attrs),
	}
}

func (h *hueHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := DefaultLogLevel
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return level >= minLevel
}

func (h *hueHandler) WithGroup(name string) slog.Handler {
	h2 := h.clone()
	h2.group = h2.group + name + "."
	return h2
}

func (h *hueHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := h.clone()

	// we want to optimise this as best as possible
	// we need to find and remove and Prefix attrs and update the current prefix
	// we then want to preformat the remaining attributes
	preBuf := buffer{}
	attrBuf := buffer{}
	for _, a := range attrs {
		if _, ok := a.Value.Any().(PrefixAttr); ok {
			h.writeStyledAttrValue(&preBuf, a, lipgloss.Style{}, false)
			preBuf.WriteString(".")
		} else {
			h.writeAttr(&attrBuf, a, h.group)
		}
	}

	h2.prefix.Write(preBuf)
	h2.attrs.Write(attrBuf)

	return h2
}

func (h *hueHandler) Handle(ctx context.Context, rec slog.Record) error {
	buf := &buffer{}

	// write time
	if !rec.Time.IsZero() {
		rec.Time.Round(0)
		h.writeTime(buf, rec.Time)
	}

	// write level
	h.writeLevel(buf, rec.Level)

	// write caller
	if h.opts.WithCaller {
		h.writeCaller(buf, rec.PC)
	}

	// write prefix
	if h.opts.WithPrefix {
		h.writePrefix(buf)
	}

	// write message
	buf.WriteString(rec.Message)
	buf.WriteString(" ")

	// write attributes
	buf.Write(h.attrs)
	rec.Attrs(func(a slog.Attr) bool {
		h.writeAttr(buf, a, h.group)
		return true
	})

	buf.WriteString("\n")

	h.mx.Lock()
	defer h.mx.Unlock()

	_, err := h.w.Write(*buf)
	return err
}

func (h *hueHandler) writeTime(buf *buffer, t time.Time) {
	if h.opts.ReplaceAttr != nil {
		attr := h.opts.ReplaceAttr(nil, slog.Time(slog.TimeKey, t))
		if attr.Value.Kind() == slog.KindTime {
			t = attr.Value.Time()
		} else {
			h.writeAttr(buf, attr, "")
			return
		}
	}

	buf.WriteString(mutedStyle.Render(t.Format(h.opts.TimeFormat)))
	buf.WriteString(" ")
}

func (h *hueHandler) writeLevel(buf *buffer, level slog.Level) {
	if h.opts.ReplaceAttr != nil {
		attr := h.opts.ReplaceAttr(nil, slog.Any(slog.LevelKey, level))
		buf.WriteString(attr.Value.String() + " ")
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

	buf.WriteString(" ")
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

	if h.opts.ReplaceAttr != nil {
		attr := h.opts.ReplaceAttr(nil, slog.Any("caller", &src))
		if v, ok := attr.Value.Any().(*slog.Source); ok {
			src = *v
		}
	}

	_, file := filepath.Split(src.File)

	// write the caller
	buf.WriteString(mutedStyle.Render(fmt.Sprintf("<%s:%d>", file, src.Line)))
	buf.WriteString(" ")
}

// writePrefix writes the prefix to the buffer, replacing the last character (.) with a colon and space.
func (h *hueHandler) writePrefix(buf *buffer) {
	if len(h.prefix) == 0 {
		return
	}

	buf.WriteString(prefixStyle.Render(string(h.prefix[:len(h.prefix)-1]) + ": "))
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

	style, found := h.attrStyle(attr, attrStyle)

	h.writeAttrKey(buf, attr, style.Faint(true), prefix)
	if !found {
		// reset the style to default if not specified by the attribute
		style = lipgloss.NewStyle()
	}
	h.writeAttrValue(buf, attr, style)
	buf.WriteString(" ")
}

func (h *hueHandler) attrStyle(attr slog.Attr, defaultStyle lipgloss.Style) (lipgloss.Style, bool) {
	if styledVal, ok := attr.Value.Any().(StyledAttr); ok {
		return styledVal.Style(), true
	}

	return defaultStyle, false
}

func (h *hueHandler) writeAttrKey(buf *buffer, attr slog.Attr, style lipgloss.Style, prefix string) {
	buf.WriteString(style.Render(fmt.Sprintf("%s=", prefix+attr.Key)))
}

func (h *hueHandler) writeAttrValue(buf *buffer, attr slog.Attr, style lipgloss.Style) {
	h.writeStyledAttrValue(buf, attr, style, true)
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
