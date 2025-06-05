# Hue Logger

Hue Logger is a Go package that provides a custom log handler for pretty formatted logs using the `log/slog` package and charmbraclets `lipgloss` for styling.

## Installation

To install the package, run:

```sh
go get github.com/smc13/hue
```

## Usage

```go
logger := slog.New(hue.NewHueHandler(os.Stderr, &hue.Options{
  Level: slog.LevelDebug,
  WithPrefix: true,
  WithCaller: true,
}))
```

### Custom attribute styling

You can provide custom styling for your attribute keys by implementing the
`hue.StyledAttr` interface:

```go
type serviceAttr string

// Style implements the hue.StyledAttr interface
func (s serviceAttr) Style() lipgloss.Style {
	return serviceAttrStyle
}

func Service(name string) slog.Attr {
	return slog.Any(ServiceKey, serviceAttr(name))
}
```

### Prefixing logs

You can prefix your logs with a custom string by passing an `slog.Attr` that implmements the
`hue.PrefixAttr` interface to `logger.With`:

```go
type PrefixAttr string
func (p PrefixAttr) Prefix() bool { return true } // identifies the attribute as a prefix

func Prefix(name string) slog.Attr {
	return slog.Any("prefix", PrefixAttr(name))
}

logger := logger.With(Prefix("my-service"))
```

Multiple prefixes will be concatenated with a dot.

For convenience, `hue.Service` already implements the `hue.PrefixAttr` interface.
