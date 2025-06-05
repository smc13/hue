package hue

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0"))
	attrStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	debugLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	infoLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	warnLevelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	errorLevelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	errorAttrStyle = errorLevelStyle
	prefixStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
)
