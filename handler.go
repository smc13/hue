package hue

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

const DefaultLogLevel = slog.LevelInfo
const DefaultTimeFormat = time.TimeOnly

type hueHandler struct {
	w  io.Writer
	mx *sync.Mutex

	opts Options

	group  string
	groups []string

	scope string
	attrs []attrNode

	colorProfile colorprofile.Profile
	width        int
}

func New(w io.Writer, options *Options) *hueHandler {
	opts := defaultOptions()
	if options != nil {
		opts.mergeFrom(options)
	}

	h := &hueHandler{
		w:    w,
		mx:   &sync.Mutex{},
		opts: opts,
	}

	h.colorProfile = detectColorProfile(w, h.opts.Color)
	h.width = h.opts.Width
	if h.width == 0 {
		h.width = detectWidth(w)
	}

	return h
}

// defaultOptions returns hue's default Options.
func defaultOptions() Options {
	return Options{
		Level:      DefaultLogLevel,
		TimeFormat: DefaultTimeFormat,
		AddScope:   true,
		AddSource:  false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
		Styles:     DefaultStyles(),
		SourceLink: FileSourceLink,
		View:       NewCompactView(),
	}
}

// mergeFrom overlays the fields o explicitly sets onto opts, leaving opts'
// defaults in place for anything o leaves at its zero value.
func (opts *Options) mergeFrom(o *Options) {
	if o.Level != nil {
		opts.Level = o.Level
	}

	if o.TimeFormat != "" {
		opts.TimeFormat = o.TimeFormat
	}

	if o.ReplaceAttr != nil {
		opts.ReplaceAttr = o.ReplaceAttr
	}

	if o.SourceLink != nil {
		opts.SourceLink = o.SourceLink
	}

	if o.Styles != nil {
		opts.Styles = o.Styles
	}

	if o.View != nil {
		opts.View = o.View
	}

	opts.Color = o.Color
	opts.Width = o.Width
	opts.AddSource = o.AddSource
	opts.AddScope = o.AddScope
}

// detectColorProfile resolves the color profile to render with, honoring an
// explicit ColorMode override, or auto-detecting from the destination writer
// and environment (respecting NO_COLOR/CLICOLOR/CLICOLOR_FORCE) otherwise.
func detectColorProfile(w io.Writer, mode ColorMode) colorprofile.Profile {
	switch mode {
	case ColorAlways:
		return colorprofile.TrueColor
	case ColorNever:
		return colorprofile.NoTTY
	default:
		return colorprofile.Detect(w, os.Environ())
	}
}

// detectWidth auto-detects the terminal width of w, returning 0 (unbounded)
// if w isn't a terminal.
func detectWidth(w io.Writer) int {
	f, ok := w.(interface{ Fd() uintptr })
	if !ok || !term.IsTerminal(f.Fd()) {
		return 0
	}

	width, _, err := term.GetSize(f.Fd())
	if err != nil {
		return 0
	}

	return width
}

func (h *hueHandler) clone() *hueHandler {
	return &hueHandler{
		w:  h.w,
		mx: h.mx,

		opts: h.opts,

		group:  h.group,
		groups: h.groups,
		scope:  h.scope,
		attrs:  slices.Clip(h.attrs),

		colorProfile: h.colorProfile,
		width:        h.width,
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
	// we need to find and remove any Scope attrs and update the current scope
	// we then want to preformat the remaining attributes
	scope := strings.Builder{}
	nodes := make([]attrNode, 0, len(attrs))
	for _, a := range attrs {
		if _, ok := a.Value.Any().(ScopeAttr); ok && h.opts.AddScope {
			text, _ := attrValueText(a.Value.Resolve())
			scope.Write([]byte(text + "."))
			continue
		}

		if node, ok := h.buildAttrNode(a, h.groups); ok {
			nodes = append(nodes, node)
		}
	}

	h2.scope = scope.String()
	h2.attrs = append(slices.Clip(h.attrs), nodes...)

	return h2
}

func (h *hueHandler) Handle(ctx context.Context, rec slog.Record) error {
	line := Line{
		Level:   rec.Level,
		Time:    rec.Time,
		HasTime: !rec.Time.IsZero(),
		Message: rec.Message,
	}

	if h.opts.AddScope {
		line.Scope = strings.TrimSuffix(h.scope, ".")
	}

	attrs := make([]attrNode, 0, len(h.attrs)+rec.NumAttrs()+1)

	if h.opts.AddSource {
		if node, ok := h.sourceAttrNode(rec); ok {
			attrs = append(attrs, node)
		}
	}

	attrs = append(attrs, h.attrs...)

	rec.Attrs(func(a slog.Attr) bool {
		if node, ok := h.buildAttrNode(a, h.groups); ok {
			attrs = append(attrs, node)
		}
		return true
	})

	line.Attrs = attrs

	buf := &buffer{}
	h.opts.View.Render(buf, renderContext{
		Styles:     h.opts.Styles,
		Width:      h.width,
		TimeFormat: h.opts.TimeFormat,
	}, line)

	h.mx.Lock()
	defer h.mx.Unlock()

	out := &colorprofile.Writer{Forward: h.w, Profile: h.colorProfile}
	_, err := out.Write(*buf)
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

// sourceAttrNode builds the "source" attribute for rec, when AddSource is
// enabled. ok is false when there's no source file to report.
func (h *hueHandler) sourceAttrNode(rec slog.Record) (attrNode, bool) {
	src := h.getSource(rec)
	if src == nil {
		src = &slog.Source{}
	}

	_, file := filepath.Split(src.File)
	if file == "" {
		return attrNode{}, false
	}

	var link string
	if h.opts.SourceLink != nil {
		link = h.opts.SourceLink(src)
	}

	text := fmt.Sprintf("%s:%d", file, src.Line)
	if link != "" {
		text = hyperlink(link, text)
	}

	return attrNode{
		key:   "source",
		value: h.opts.Styles.Source.Render(text),
	}, true
}

// hyperlink creates a terminal-friendly hyperlink using the OSC 8 escape sequence.
func hyperlink(url, label string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, label)
}
