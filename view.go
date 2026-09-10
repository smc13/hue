package hue

import (
	"log/slog"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"
)

// Line is the structured representation of a single log line, gathered by
// the handler and handed off to a View for rendering.
type Line struct {
	Level   slog.Level
	Time    time.Time
	HasTime bool
	Scope   string
	Message string
	Attrs   []attrNode
}

// renderContext carries the shared configuration a View needs to render a
// Line: styling, wrap width, and time formatting.
type renderContext struct {
	Styles     *Styles
	Width      int // 0 = unbounded (no wrapping)
	TimeFormat string
}

// View renders a Line to buf. Views are instantiated once per handler and
// reused for every log line.
type View interface {
	Render(buf *buffer, ctx renderContext, line Line)
}

// renderHead renders the shared level/time/scope/separator prefix common to
// every view (eg. "DBG 12:00:00 my-scope › "), returning the rendered text
// and its display width.
func renderHead(ctx renderContext, line Line) (string, int) {
	var head buffer

	levelStyle := ctx.Styles.levelStyle(line.Level)
	sepStyle := lipgloss.NewStyle().Foreground(levelStyle.GetForeground())

	head.WriteString(levelStyle.String())
	head.WriteString(" ")

	if line.HasTime {
		head.WriteString(ctx.Styles.Time.Render(line.Time.Format(ctx.TimeFormat)))
		head.WriteString(" ")
	}

	if line.Scope != "" {
		head.WriteString(ctx.Styles.Scope.Render(line.Scope))
		head.WriteString(" ")
	}

	head.WriteString(sepStyle.Render("\u203a"))
	head.WriteString(" ")

	return string(head), lipgloss.Width(string(head))
}

// writeMainLine writes the line shared by CompactView and TreeView:
//
//	<level> <time> <scope> › <message>
//
// Long messages wrap with continuation lines indented to align under the
// message's start column.
func writeMainLine(buf *buffer, ctx renderContext, line Line) {
	head, indent := renderHead(ctx, line)

	message := line.Message
	if avail := ctx.Width - indent; ctx.Width > 0 && avail > 0 {
		message = lipgloss.Wrap(message, avail, " ")
	}
	message = indentContinuation(message, indent)

	buf.WriteString(head)
	buf.WriteString(message)
	buf.WriteString("\n")
}

// attrIndentWidth returns the small, fixed indent used to set off
// attributes (CompactView's attribute line, TreeView's tree) from the left
// margin: the width of the level text plus one space, mirroring the gap
// between the level and the time in the main line.
func attrIndentWidth(ctx renderContext, level slog.Level) int {
	return lipgloss.Width(ctx.Styles.levelStyle(level).Value())
}

// indentContinuation prepends width spaces to every line after the first.
func indentContinuation(s string, width int) string {
	if !strings.Contains(s, "\n") {
		return s
	}

	pad := strings.Repeat(" ", width)
	lines := strings.Split(s, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = pad + lines[i]
	}

	return strings.Join(lines, "\n")
}

// indentAllLines prepends width spaces to every line of s, including the first.
func indentAllLines(s string, width int) string {
	if width <= 0 {
		return s
	}

	pad := strings.Repeat(" ", width)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = pad + lines[i]
	}

	return strings.Join(lines, "\n")
}

// CompactView renders attributes on their own line(s), wrapped and indented
// to align under the message.
type CompactView struct{}

// NewCompactView creates a new CompactView.
func NewCompactView() *CompactView { return &CompactView{} }

func (v *CompactView) Render(buf *buffer, ctx renderContext, line Line) {
	writeMainLine(buf, ctx, line)
	if len(line.Attrs) == 0 {
		return
	}

	flat := flattenAttrNodes(line.Attrs, "", ctx.Styles.Key)
	if len(flat) == 0 {
		return
	}

	indent := attrIndentWidth(ctx, line.Level)

	attrLine := strings.Join(flat, " ")
	if avail := ctx.Width - indent; ctx.Width > 0 && avail > 0 {
		attrLine = lipgloss.Wrap(attrLine, avail, " ")
	}
	attrLine = indentContinuation(attrLine, indent)

	buf.WriteString(strings.Repeat(" ", indent))
	buf.WriteString(attrLine)
	buf.WriteString("\n")
}

// TreeView renders attributes as a tree beneath the main line, with nested
// groups rendered as nested branches.
type TreeView struct{}

// NewTreeView creates a new TreeView.
func NewTreeView() *TreeView { return &TreeView{} }

func (v *TreeView) Render(buf *buffer, ctx renderContext, line Line) {
	writeMainLine(buf, ctx, line)
	if len(line.Attrs) == 0 {
		return
	}

	enumStyle := ctx.Styles.Key.MarginRight(1)
	t := tree.Root("").EnumeratorStyle(enumStyle).IndenterStyle(ctx.Styles.Key)
	appendAttrNodes(t, line.Attrs, ctx.Styles.Key, enumStyle)

	indent := attrIndentWidth(ctx, line.Level)
	buf.WriteString(indentAllLines(t.String(), indent))
	buf.WriteString("\n")
}

// appendAttrNodes adds attrNodes as children of t, recursing into nested
// groups as nested trees.
func appendAttrNodes(t *tree.Tree, nodes []attrNode, keyStyle, enumStyle lipgloss.Style) {
	for _, n := range nodes {
		if n.isGroup {
			child := tree.Root(keyStyle.Render(n.key)).EnumeratorStyle(enumStyle).IndenterStyle(keyStyle)
			appendAttrNodes(child, n.children, keyStyle, enumStyle)
			t.Child(child)
			continue
		}

		t.Child(keyStyle.Render(n.key+":") + " " + n.value)
	}
}

// SingleLineView renders the level, time, scope, message, and attributes all on a single line.
type SingleLineView struct{}

// NewSingleLineView creates a new SingleLineView.
func NewSingleLineView() *SingleLineView { return &SingleLineView{} }

func (v *SingleLineView) Render(buf *buffer, ctx renderContext, line Line) {
	head, _ := renderHead(ctx, line)

	buf.WriteString(head)
	buf.WriteString(line.Message)

	if flat := flattenAttrNodes(line.Attrs, "", ctx.Styles.Key); len(flat) > 0 {
		buf.WriteString(" ")
		buf.WriteString(strings.Join(flat, " "))
	}

	buf.WriteString("\n")
}
