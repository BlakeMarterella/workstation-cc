// Package ui holds shared Lip Gloss styles for colored, readable output. Lip
// Gloss automatically degrades color when stdout is not a capable terminal, so
// the same styles are safe in piped / non-interactive contexts.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Title styles a top-level heading.
	Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	// Success styles positive results.
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	// Warn styles cautionary text.
	Warn = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	// Faint styles secondary/detail text.
	Faint = lipgloss.NewStyle().Faint(true)
)

// Header renders a section heading with a leading marker.
func Header(s string) string {
	return Title.Render("==> " + s)
}
