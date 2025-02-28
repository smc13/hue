package hue

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	attrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	debugLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	infoLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	warnLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	errorLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	errorAttrStyle   = errorLevelStyle.Faint(true)
	serviceAttrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("219"))
)
