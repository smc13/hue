# Hue Logger

Hue Logger is a Go package that provides a custom `log/slog` handler for
pretty, colourful log output, styled using charmbracelet's `lipgloss` (v2).

## Installation

To install the package, run:

```sh
go get github.com/smc13/hue
```

## Usage

```go
logger := slog.New(hue.New(os.Stderr, &hue.Options{
	Level:     slog.LevelDebug,
	AddScope:  true,
	AddSource: true,
}))
```

Every log line follows the same format:

```
<level> <time> <scope> › <message>
```

The level and the `›` separator are coloured according to the log level.
Long messages word-wrap, with continuation lines indented to align under the
start of the message.

## Views

Attributes are rendered by a pluggable `hue.View`, configured once via
`Options.View` at handler setup:

- `hue.NewCompactView()` (the default) — attributes are flattened onto the
  line(s) below the message, wrapped and indented under a small, fixed
  margin.
- `hue.NewTreeView()` — attributes are rendered as a tree beneath the main
  line, with nested `slog.Group`s rendered as nested branches.
- `hue.NewSingleLineView()` — the level, time, scope, message, and
  attributes are all rendered on a single line with no wrapping, matching
  hue's original output format.

CompactView and TreeView indent their attributes by a small, fixed amount
(the width of the level text, mirroring the gap between the level and the
time), rather than aligning under the message.

```go
logger := slog.New(hue.New(os.Stderr, &hue.Options{
	View: hue.NewTreeView(),
}))
```

```
INF 12:00:00 auth-service › user authenticated
   ├── userID: 123
   ├── requestID: abc-123
   └── testing
      ├── now: "2026-09-02 12:00:00.123456 +0930 ACST"
      └── duration: 5s
```

```go
logger := slog.New(hue.New(os.Stderr, &hue.Options{
	View: hue.NewSingleLineView(),
}))
```

```
INF 12:00:00 auth-service › user authenticated userID=123 requestID=abc-123 testing.now="2026-09-02 12:00:00.123456 +0930 ACST" testing.duration=5s
```

## Themes

Hue ships with a few built-in `Styles` presets, set via `Options.Styles`:

- `hue.DefaultStyles()` (the default) — hue's original ANSI-based theme.
- `hue.CatppuccinLatteStyles()` — [Catppuccin](https://catppuccin.com) Latte (light).
- `hue.CatppuccinFrappeStyles()` — Catppuccin Frappé.
- `hue.CatppuccinMacchiatoStyles()` — Catppuccin Macchiato.
- `hue.CatppuccinMochaStyles()` — Catppuccin Mocha.

The Catppuccin themes map hue's roles onto the palette following the
[Catppuccin style guide](https://github.com/catppuccin/catppuccin/blob/main/docs/style-guide.md):
Text for body copy, Subtext 0 for labels/time, Overlay 1 for debug, Blue for
info, Yellow for warnings, Red for errors, Teal for scope, Mauve for booleans,
Peach for numbers, and Pink for durations/times.

```go
logger := slog.New(hue.New(os.Stderr, &hue.Options{
	Styles: hue.CatppuccinMochaStyles(),
}))
```

## Color and width

- `Options.Color` controls whether/how output is coloured: `hue.ColorAuto`
  (default, detects based on the destination and environment, respecting
  `NO_COLOR`/`CLICOLOR`/`CLICOLOR_FORCE`), `hue.ColorAlways`, or
  `hue.ColorNever`.
- `Options.Width` sets a fixed wrap width for messages and attributes. It
  defaults to `0`, which auto-detects the width of the destination terminal,
  falling back to unbounded (no wrapping) when the destination isn't a
  terminal.

## Attribute value styling

Attribute values are coloured based on their `slog.Kind` by default (booleans,
numbers, durations, and times each get a distinct colour from
`Options.Styles.Values`), with plain strings and everything else falling back
to `Options.Styles.Attr`. Values are only quoted when necessary to stay
unambiguous — for example, an empty string or a string containing whitespace,
quotes, or `=`.

You can override the styling for any of these value kinds, or provide a
per-attribute style that always takes precedence, by implementing the
`hue.StyledAttr` interface:

```go
type errorAttr struct{ error }

// Style implements the hue.StyledAttr interface
func (e errorAttr) Style() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
}
```

`hue.Err` already implements this to style errors.

## Scoping logs

You can scope your logs with a custom string by passing an `slog.Attr` that
implements the `hue.ScopeAttr` interface to `logger.With`:

```go
type myScopeAttr string
func (s myScopeAttr) Prefix() bool { return true } // identifies the attribute as a scope

func MyScope(name string) slog.Attr {
	return slog.Any("scope", myScopeAttr(name))
}

logger := logger.With(MyScope("my-service"))
```

Multiple scopes will be concatenated with a dot (eg. `my-service.sub`).

For convenience, `hue.Scope` already implements the `hue.ScopeAttr` interface
(and, via `hue.StyledAttr`, is styled consistently too):

```go
logger := logger.With(hue.Scope("my-service"))
```

> [!NOTE]
> `hue.Service`, `hue.ServiceKey`, and `hue.PrefixAttr` are deprecated aliases
> for `hue.Scope`, `hue.ScopeKey`, and `hue.ScopeAttr` respectively, kept for
> backwards compatibility.

## Source location

When `Options.AddSource` is enabled, a `source` attribute (eg.
`source=main.go:42`) is added to every log line. `Options.SourceLink` can be
used to make it a clickable terminal hyperlink, eg:

```go
logger := slog.New(hue.New(os.Stderr, &hue.Options{
	AddSource:  true,
	SourceLink: hue.VscodeSourceLink, // or hue.FileSourceLink
}))
```
