package hue

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	attrStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	debugLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	infoLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("155"))
	warnLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	errorLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	errorAttrStyle   = errorLevelStyle
	serviceAttrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("219"))
)
