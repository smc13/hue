package hue

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	attrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	debugLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	infoLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("123"))
	warnLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	errorLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	errorAttrStyle   = errorLevelStyle.Faint(true)
	serviceAttrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("219"))
)
