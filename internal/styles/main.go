package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	PrimaryColor    = lipgloss.Color("#7C3AED")
	SecondaryColor  = lipgloss.Color("#A855F7")
	SuccessColor    = lipgloss.Color("#10B981")
	ErrorColor      = lipgloss.Color("#EF4444")
	WarningColor    = lipgloss.Color("#F59E0B")
	MutedColor      = lipgloss.Color("#6B7280")
	BackgroundColor = lipgloss.Color("#1F2937")
	TextColor       = lipgloss.Color("#F9FAFB")
	BorderColor     = lipgloss.Color("#374151")
)

// Common styles
var (
	// Typography
	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Align(lipgloss.Center)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Background(PrimaryColor).
			Bold(true).
			Padding(0, 2)

	SubheaderStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Layout styles
	ContentStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			Padding(1)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2).
			MarginBottom(1)

	// Interactive styles
	ActiveStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	InactiveStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Background(lipgloss.Color("#312E81"))

	// Progress styles
	ProgressStyle = lipgloss.NewStyle().
			MarginBottom(1)

	ProgressBarStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(BorderColor).
				Padding(0, 1)

	// List styles
	ListStyle = lipgloss.NewStyle().
			Padding(1, 0).
			MarginTop(1)

	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			MarginBottom(0)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Align(lipgloss.Center)

	// Container styles
	ContainerStyle = lipgloss.NewStyle().
			Padding(2, 4).
			MaxWidth(80)

	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2).
			Width(20)

	MainContentStyle = lipgloss.NewStyle().
				Padding(1, 2)
)

// Step-specific styles
var (
	StepIndicatorStyle = lipgloss.NewStyle().
				Foreground(MutedColor).
				MarginBottom(1)

	CompletedStepStyle = lipgloss.NewStyle().
				Foreground(SuccessColor)

	ActiveStepStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	PendingStepStyle = lipgloss.NewStyle().
				Foreground(MutedColor)
)

// Package-specific styles for install step
var (
	PackageListStyle = lipgloss.NewStyle().
				Padding(1, 0).
				MarginTop(1)

	CompletedPackageStyle = lipgloss.NewStyle().
				Foreground(SuccessColor).
				PaddingLeft(2)

	ActivePackageStyle = lipgloss.NewStyle().
				Foreground(WarningColor).
				Bold(true).
				PaddingLeft(2)

	PendingPackageStyle = lipgloss.NewStyle().
				Foreground(MutedColor).
				PaddingLeft(2)
)

// Utility functions
func WithMaxWidth(style lipgloss.Style, width int) lipgloss.Style {
	return style.MaxWidth(width)
}

func WithPadding(style lipgloss.Style, vertical, horizontal int) lipgloss.Style {
	return style.Padding(vertical, horizontal)
}

func WithMargin(style lipgloss.Style, vertical, horizontal int) lipgloss.Style {
	return style.Margin(vertical, horizontal)
}

func WithBorder(style lipgloss.Style, border lipgloss.Border, color lipgloss.Color) lipgloss.Style {
	return style.Border(border).BorderForeground(color)
}

// Theme variants
func GetTheme(name string) map[string]lipgloss.Color {
	themes := map[string]map[string]lipgloss.Color{
		"default": {
			"primary":    PrimaryColor,
			"secondary":  SecondaryColor,
			"success":    SuccessColor,
			"error":      ErrorColor,
			"warning":    WarningColor,
			"muted":      MutedColor,
			"background": BackgroundColor,
			"text":       TextColor,
			"border":     BorderColor,
		},
		"dark": {
			"primary":    lipgloss.Color("#8B5CF6"),
			"secondary":  lipgloss.Color("#A78BFA"),
			"success":    lipgloss.Color("#34D399"),
			"error":      lipgloss.Color("#F87171"),
			"warning":    lipgloss.Color("#FBBF24"),
			"muted":      lipgloss.Color("#9CA3AF"),
			"background": lipgloss.Color("#111827"),
			"text":       lipgloss.Color("#F9FAFB"),
			"border":     lipgloss.Color("#374151"),
		},
		"light": {
			"primary":    lipgloss.Color("#7C3AED"),
			"secondary":  lipgloss.Color("#A855F7"),
			"success":    lipgloss.Color("#059669"),
			"error":      lipgloss.Color("#DC2626"),
			"warning":    lipgloss.Color("#D97706"),
			"muted":      lipgloss.Color("#6B7280"),
			"background": lipgloss.Color("#FFFFFF"),
			"text":       lipgloss.Color("#1F2937"),
			"border":     lipgloss.Color("#D1D5DB"),
		},
	}

	if theme, exists := themes[name]; exists {
		return theme
	}
	return themes["default"]
}
