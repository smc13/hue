package hue

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	attrStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	debugLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	infoLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("155"))
	warnLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	errorLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	errorAttrStyle   = errorLevelStyle.Faint(true)
	serviceAttrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("219"))
)
