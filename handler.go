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
	groups []string

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
			AddPrefix:  true,
			AddSource:  false,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				return a
			},
			Styles:     DefaultStyles(),
			SourceLink: FileSourceLink,
		},
	}

	if options != nil {
		h.opts = *options
		if h.opts.Styles == nil {
			h.opts.Styles = DefaultStyles()
		}
	}

	return h
}

func (h *hueHandler) clone() *hueHandler {
	return &hueHandler{
		w:  h.w,
		mx: h.mx,

		opts: h.opts,

		group:  h.group,
		groups: h.groups,
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
	h2.group += name + "."
	h2.groups = append(h2.groups, name)

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
		if _, ok := a.Value.Any().(PrefixAttr); ok && h.opts.AddPrefix {
			h.writeStyledAttrValue(&preBuf, a, lipgloss.Style{}, false)
			preBuf.WriteString(".")
		} else {
			h.writeAttr(&attrBuf, a, h.group, h.groups)
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
		h.writeTime(buf, rec.Time)
	}

	// write level
	h.writeLevel(buf, rec.Level)

	// write caller
	if h.opts.AddSource {
		src := h.getSource(rec)
		if src == nil {
			src = &slog.Source{}
		}

		h.writeSource(buf, src)
	}

	// write prefix
	if h.opts.AddPrefix {
		h.writePrefix(buf)
	}

	// write message
	buf.WriteString(rec.Message)
	buf.WriteString(" ")

	// write attributes
	h.writeAttrs(buf, rec)

	buf.WriteString("\n")

	h.mx.Lock()
	defer h.mx.Unlock()

	_, err := h.w.Write(*buf)
	return err
}

// It appears future versions of Go will expose slog.Record.Source()
// but for now we replicate its basic functionality here.
func (h *hueHandler) getSource(rec slog.Record) *slog.Source {
	// grab the caller from the stack
	frames := runtime.CallersFrames([]uintptr{rec.PC})
	frame, _ := frames.Next()

	src := &slog.Source{
		Function: frame.Function,
		File:     frame.File,
		Line:     frame.Line,
	}

	if h.opts.ReplaceAttr != nil {
		attr := h.opts.ReplaceAttr(nil, slog.Any(slog.SourceKey, &src))
		if v, ok := attr.Value.Any().(*slog.Source); ok {
			src = v
		}
	}

	return src
}

func (h *hueHandler) writeTime(buf *buffer, t time.Time) {
	buf.WriteString(h.opts.Styles.Time.Render(t.Format(h.opts.TimeFormat)))
	buf.WriteString(" ")
}

func (h *hueHandler) writeLevel(buf *buffer, level slog.Level) {
	var style lipgloss.Style
	if s, ok := h.opts.Styles.Levels[level]; ok {
		style = s
	} else {
		style = h.opts.Styles.Attr.SetString(level.String())
	}

	buf.WriteString(style.String())
	buf.WriteString(" ")
}

func (h *hueHandler) writeSource(buf *buffer, src *slog.Source) {
	_, file := filepath.Split(src.File)
	if file == "" {
		return
	}

	var link string
	if h.opts.SourceLink != nil {
		link = h.opts.SourceLink(src)
	}

	text := fmt.Sprintf("<%s:%d>", file, src.Line)

	if link == "" {
		buf.WriteString(h.opts.Styles.Source.Render(text))
	} else {
		buf.WriteString(h.opts.Styles.Source.Render(hyperlink(link, text)))
	}

	buf.WriteString(" ")
}

// hyperlink creates a terminal-friendly hyperlink using the OSC 8 escape sequence.
func hyperlink(url, label string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, label)
}

// writePrefix writes the prefix to the buffer, replacing the last character (.) with a space.
func (h *hueHandler) writePrefix(buf *buffer) {
	if len(h.prefix) == 0 {
		return
	}

	buf.WriteString(h.opts.Styles.Prefix.Render(string(h.prefix[:len(h.prefix)-1]) + " "))
}

func (h *hueHandler) writeAttrs(buf *buffer, rec slog.Record) {
	// write the pre formatted attributes
	if len(h.attrs) > 0 {
		buf.Write(h.attrs)
	}

	rec.Attrs(func(a slog.Attr) bool {
		h.writeAttr(buf, a, h.group, h.groups)
		return true
	})
}

func (h *hueHandler) writeAttr(buf *buffer, attr slog.Attr, prefix string, groups []string) {
	if rep := h.opts.ReplaceAttr; rep != nil {
		attr = rep(groups, attr)
	}

	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			prefix = prefix + attr.Key + "."
			groups = append(groups, attr.Key)
		}

		for _, groupAttrs := range attr.Value.Group() {
			h.writeAttr(buf, groupAttrs, prefix, groups)
		}

		return
	}

	var style lipgloss.Style
	var found bool
	attr.Value, style, found = h.attrStyle(attr)

	h.writeAttrKey(buf, attr, style.Faint(true), prefix)
	if !found {
		// reset the style to default if not specified by the attribute
		style = lipgloss.NewStyle()
	}
	h.writeAttrValue(buf, attr, style)
	buf.WriteString(" ")
}

func (h *hueHandler) attrStyle(attr slog.Attr) (slog.Value, lipgloss.Style, bool) {
	res := attr.Value.Resolve()

	if styledVal, ok := attr.Value.Any().(StyledAttr); ok {
		return res, styledVal.Style(), true
	}

	return res, h.opts.Styles.Attr, false
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
