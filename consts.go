package hue

import (
	"log/slog"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	DefaultTimeFormat = time.TimeOnly
	DefaultLogLevel   = slog.LevelInfo
)

var (
	timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	attrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	debugLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	infoLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("123"))
	warnLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	errorLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)
