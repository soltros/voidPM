package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Void Linux Brand Palette
	PrimaryColor   = lipgloss.Color("#478061") // Void Linux Green
	AccentColor    = lipgloss.Color("#2ecc71") // Bright Emerald
	SecondaryColor = lipgloss.Color("#3498db") // Void Blue
	WarningColor   = lipgloss.Color("#f1c40f") // Yellow
	ErrorColor     = lipgloss.Color("#e74c3c") // Red
	MutedColor     = lipgloss.Color("#7f8c8d") // Gray
	BgDark         = lipgloss.Color("#1e222a") // Dark slate background
	FgLight        = lipgloss.Color("#f8f9fa") // Off white text

	// Lipgloss Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor).
			MarginBottom(1)

	BannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(PrimaryColor).
			Padding(0, 1).
			MarginBottom(1)

	SubTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor)

	RunningBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#27ae60")).
			Padding(0, 1)

	StoppedBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(WarningColor).
			Padding(0, 1)

	DisabledBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(MutedColor).
			Padding(0, 1)

	ErrorBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ErrorColor).
			Padding(0, 1)

	InstalledBadge = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	OrphanBadge = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	HoldBadge = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)
)

func PrintBanner() {
	fmt.Println(BannerStyle.Render(" voidPM ") + " " + SubTitleStyle.Render("Void Linux System & Package Manager Helper"))
}

func RenderHeader(title string) string {
	return TitleStyle.Render("==> " + title)
}

func RenderSuccess(msg string) string {
	return lipgloss.NewStyle().Foreground(AccentColor).Render("[OK] " + msg)
}

func RenderWarning(msg string) string {
	return lipgloss.NewStyle().Foreground(WarningColor).Render("[WARN] " + msg)
}

func RenderError(msg string) string {
	return lipgloss.NewStyle().Foreground(ErrorColor).Render("[ERROR] " + msg)
}

func RenderInfo(msg string) string {
	return lipgloss.NewStyle().Foreground(SecondaryColor).Render("[INFO] " + msg)
}

// Table helper for formatted console output
type Column struct {
	Title string
	Width int
}

func RenderTable(cols []Column, rows [][]string) string {
	var sb strings.Builder

	// Header row
	var headers []string
	for _, col := range cols {
		hdr := lipgloss.NewStyle().Bold(true).Foreground(SecondaryColor).Render(padRight(col.Title, col.Width))
		headers = append(headers, hdr)
	}
	sb.WriteString(strings.Join(headers, "  ") + "\n")

	// Separator
	var seps []string
	for _, col := range cols {
		seps = append(seps, strings.Repeat("─", col.Width))
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(MutedColor).Render(strings.Join(seps, "  ")) + "\n")

	// Rows
	for _, row := range rows {
		var cells []string
		for i, cell := range row {
			w := cols[i].Width
			cells = append(cells, padRight(cell, w))
		}
		sb.WriteString(strings.Join(cells, "  ") + "\n")
	}

	return sb.String()
}

func padRight(str string, length int) string {
	plainLen := lipgloss.Width(str)
	if plainLen >= length {
		return str
	}
	return str + strings.Repeat(" ", length-plainLen)
}
