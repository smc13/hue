package hue

import (
	"encoding"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// attrNode is a structured representation of a single slog.Attr (or a
// slog.KindGroup of them), used so that views can render attributes either
// flat (CompactView) or as a tree (TreeView) from the same underlying data.
//
// Leaf nodes carry an already-styled/quoted value string; group nodes carry
// no value and instead hold their children.
type attrNode struct {
	key      string
	value    string
	isGroup  bool
	children []attrNode
}

// buildAttrNode resolves and styles a single slog.Attr, returning ok=false
// if the attribute should be skipped entirely (eg. removed by ReplaceAttr).
func (h *hueHandler) buildAttrNode(attr slog.Attr, groups []string) (attrNode, bool) {
	if rep := h.opts.ReplaceAttr; rep != nil {
		attr = rep(groups, attr)
	}

	if attr.Equal(slog.Attr{}) {
		return attrNode{}, false
	}

	if attr.Value.Kind() == slog.KindGroup {
		childGroups := groups
		if attr.Key != "" {
			childGroups = append(append([]string{}, groups...), attr.Key)
		}

		children := make([]attrNode, 0, len(attr.Value.Group()))
		for _, ga := range attr.Value.Group() {
			if node, ok := h.buildAttrNode(ga, childGroups); ok {
				children = append(children, node)
			}
		}

		if len(children) == 0 {
			return attrNode{}, false
		}

		return attrNode{key: attr.Key, isGroup: true, children: children}, true
	}

	val, style := h.attrValueStyle(attr)
	return attrNode{
		key:   attr.Key,
		value: formatAttrValue(val, style),
	}, true
}

// attrValueStyle resolves the style used to render an attribute's value.
// StyledAttr always takes precedence; otherwise a default is chosen based on
// the value's kind, falling back to the generic Attr style.
func (h *hueHandler) attrValueStyle(attr slog.Attr) (slog.Value, lipgloss.Style) {
	res := attr.Value.Resolve()

	if styledVal, ok := attr.Value.Any().(StyledAttr); ok {
		return res, styledVal.Style()
	}

	switch res.Kind() {
	case slog.KindBool:
		return res, h.opts.Styles.Values.Bool
	case slog.KindInt64, slog.KindUint64, slog.KindFloat64:
		return res, h.opts.Styles.Values.Number
	case slog.KindDuration:
		return res, h.opts.Styles.Values.Duration
	case slog.KindTime:
		return res, h.opts.Styles.Values.Time
	default:
		return res, h.opts.Styles.Attr
	}
}

// attrValueText returns the plain, unstyled, unquoted text representation of
// val, along with whether that text is eligible for quoting (ie. it's
// free-form text rather than a fixed-format value like a bool or number).
func attrValueText(val slog.Value) (text string, quotable bool) {
	switch val.Kind() {
	case slog.KindString:
		return val.String(), true
	case slog.KindBool:
		return strconv.FormatBool(val.Bool()), false
	case slog.KindInt64:
		return strconv.FormatInt(val.Int64(), 10), false
	case slog.KindUint64:
		return strconv.FormatUint(val.Uint64(), 10), false
	case slog.KindFloat64:
		return strconv.FormatFloat(val.Float64(), 'f', -1, 64), false
	case slog.KindTime:
		return val.Time().String(), true
	case slog.KindDuration:
		return val.Duration().String(), false
	case slog.KindAny:
		switch avt := val.Any().(type) {
		case encoding.TextMarshaler:
			if enc, err := avt.MarshalText(); err == nil {
				return string(enc), true
			}
			return fmt.Sprintf("%+v", avt), true
		case fmt.Stringer:
			return avt.String(), true
		default:
			return fmt.Sprintf("%+v", avt), true
		}
	default:
		return "", false
	}
}

// formatAttrValue formats and styles a resolved slog.Value, applying "smart"
// quoting: values are only quoted when necessary to keep them unambiguous
// (empty, or containing whitespace/quotes/`=`).
func formatAttrValue(val slog.Value, style lipgloss.Style) string {
	text, quotable := attrValueText(val)
	if quotable {
		text = quoteIfNeeded(text)
	}
	return style.Render(text)
}

// quoteIfNeeded quotes s only when necessary to keep it unambiguous: when
// it's empty, or contains whitespace, quotes, or an `=` character.
func quoteIfNeeded(s string) string {
	if s == "" || strings.ContainsAny(s, " \t\n\"=") {
		return strconv.Quote(s)
	}
	return s
}

// flattenAttrNodes renders a tree of attrNodes into a flat list of styled
// "key=value" strings, joining nested group keys with ".".
func flattenAttrNodes(nodes []attrNode, prefix string, keyStyle lipgloss.Style) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.isGroup {
			childPrefix := prefix
			if n.key != "" {
				childPrefix = prefix + n.key + "."
			}
			out = append(out, flattenAttrNodes(n.children, childPrefix, keyStyle)...)
			continue
		}

		out = append(out, keyStyle.Render(prefix+n.key+"=")+n.value)
	}

	return out
}
