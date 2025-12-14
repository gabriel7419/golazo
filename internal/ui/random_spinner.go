package ui

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// RandomCharSpinner is a custom spinner that cycles through random characters.
type RandomCharSpinner struct {
	chars      []rune
	currentIdx int
	width      int
	startColor colorful.Color // Gradient start color (neon cyan)
	endColor   colorful.Color // Gradient end color (neon magenta - complementary)
}

// NewRandomCharSpinner creates a new random character spinner.
func NewRandomCharSpinner() *RandomCharSpinner {
	// Random characters similar to the image: alphanumeric, symbols, special chars
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+-=[]{}|;:,.<>?/~`£€¥")

	// Create neon gradient: cyan to magenta (complementary neon colors)
	startColor, _ := colorful.Hex("#00FFFF") // Bright cyan (neon accent color)
	endColor, _ := colorful.Hex("#FF00FF")   // Bright magenta (complementary neon)

	return &RandomCharSpinner{
		chars:      chars,
		currentIdx: rand.Intn(len(chars)),
		width:      20, // Default width for spinner
		startColor: startColor,
		endColor:   endColor,
	}
}

// Init initializes the spinner with a tick command.
func (r *RandomCharSpinner) Init() tea.Cmd {
	return r.tick()
}

// Model interface compatibility - not used but needed for Update signature
func (r *RandomCharSpinner) Model() tea.Model {
	return r
}

// Update updates the spinner state.
func (r *RandomCharSpinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		// Change to a random character very quickly
		r.currentIdx = rand.Intn(len(r.chars))
		return r, r.tick()
	}
	return r, nil
}

// View renders the spinner with gradient colors.
func (r *RandomCharSpinner) View() string {
	// Create a string of random characters for the spinner
	spinnerChars := make([]rune, r.width)
	for i := range spinnerChars {
		spinnerChars[i] = r.chars[rand.Intn(len(r.chars))]
	}

	// Apply gradient to each character
	var result strings.Builder
	for i, char := range spinnerChars {
		// Calculate gradient position (0.0 to 1.0)
		ratio := float64(i) / float64(r.width-1)

		// Blend colors based on position
		color := r.startColor.BlendLab(r.endColor, ratio)

		// Convert to hex for lipgloss
		hexColor := color.Hex()

		// Style each character with its gradient color
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor))
		result.WriteString(charStyle.Render(string(char)))
	}

	return result.String()
}

// SetWidth sets the width of the spinner.
func (r *RandomCharSpinner) SetWidth(width int) {
	r.width = width
}

// TickMsg is a message sent to update the spinner.
type TickMsg struct{}

// tick sends a tick message after a very short delay for fast, smooth animation.
func (r *RandomCharSpinner) tick() tea.Cmd {
	// Much faster update (20ms) for smoother animation
	return tea.Tick(20*time.Millisecond, func(time.Time) tea.Msg {
		return TickMsg{}
	})
}
