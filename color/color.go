package color

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func escape(code string) string {
	return fmt.Sprintf("\033[%sm", code)
}

func wrap(code, text string) string {
	return fmt.Sprintf("%s%s%s", escape(code), text, escape("0"))
}

func wrapEscape(code string, x string) string {
	return fmt.Sprintf("%s%s%s", escape(code), x, escape("0"))
}

// =============================================================================
// BASIC FOREGROUND COLORS (30-37)
// =============================================================================

func Black(text string) string   { return wrap("30", text) }
func Red(text string) string     { return wrap("31", text) }
func Green(text string) string   { return wrap("32", text) }
func Yellow(text string) string  { return wrap("33", text) }
func Blue(text string) string    { return wrap("34", text) }
func Magenta(text string) string { return wrap("35", text) }
func Cyan(text string) string    { return wrap("36", text) }
func White(text string) string   { return wrap("37", text) }

// =============================================================================
// BRIGHT FOREGROUND COLORS (90-97)
// =============================================================================

func BrightBlack(text string) string   { return wrap("90", text) }
func BrightRed(text string) string     { return wrap("91", text) }
func BrightGreen(text string) string   { return wrap("92", text) }
func BrightYellow(text string) string  { return wrap("93", text) }
func BrightBlue(text string) string    { return wrap("94", text) }
func BrightMagenta(text string) string { return wrap("95", text) }
func BrightCyan(text string) string    { return wrap("96", text) }
func BrightWhite(text string) string   { return wrap("97", text) }

// =============================================================================
// BACKGROUND COLORS (40-47)
// =============================================================================

func BlackBg(text string) string   { return wrap("40", text) }
func RedBg(text string) string     { return wrap("41", text) }
func GreenBg(text string) string   { return wrap("42", text) }
func YellowBg(text string) string  { return wrap("43", text) }
func BlueBg(text string) string    { return wrap("44", text) }
func MagentaBg(text string) string { return wrap("45", text) }
func CyanBg(text string) string    { return wrap("46", text) }
func WhiteBg(text string) string   { return wrap("47", text) }

// =============================================================================
// BRIGHT BACKGROUND COLORS (100-107)
// =============================================================================

func BrightBlackBg(text string) string   { return wrap("100", text) }
func BrightRedBg(text string) string     { return wrap("101", text) }
func BrightGreenBg(text string) string   { return wrap("102", text) }
func BrightYellowBg(text string) string  { return wrap("103", text) }
func BrightBlueBg(text string) string    { return wrap("104", text) }
func BrightMagentaBg(text string) string { return wrap("105", text) }
func BrightCyanBg(text string) string    { return wrap("106", text) }
func BrightWhiteBg(text string) string   { return wrap("107", text) }

// =============================================================================
// TEXT FORMATTING
// =============================================================================

func Bold(text string) string          { return wrap("1", text) }
func Dim(text string) string           { return wrap("2", text) }
func Italic(text string) string        { return wrap("3", text) }
func Underline(text string) string     { return wrap("4", text) }
func Blink(text string) string         { return wrap("5", text) }
func Reverse(text string) string       { return wrap("7", text) }
func Hidden(text string) string        { return wrap("8", text) }
func Strikethrough(text string) string { return wrap("9", text) }

// =============================================================================
// COMBINED FORMATTING (as shown in transcription)
// =============================================================================

func BoldRed(text string) string {
	return fmt.Sprintf("\033[1;31m%s\033[0m", text)
}

func BoldGreen(text string) string {
	return fmt.Sprintf("\033[1;32m%s\033[0m", text)
}

func BoldYellow(text string) string {
	return fmt.Sprintf("\033[1;33m%s\033[0m", text)
}

func BoldBlue(text string) string {
	return fmt.Sprintf("\033[1;34m%s\033[0m", text)
}

// =============================================================================
// 256-COLOR SUPPORT
// =============================================================================

// supports256Color checks if terminal supports 256 colors
func supports256Color() bool {
	term := os.Getenv("TERM")
	return strings.Contains(term, "256color") ||
		strings.Contains(term, "xterm") ||
		strings.Contains(term, "screen")
}

// Color256 sets foreground color using 256-color palette
func Color256(colorNumber int, text string) string {
	if !supports256Color() {
		// Fallback to nearest standard color
		return fallbackColor(colorNumber, text)
	}
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", colorNumber, text)
}

// Background256 sets background color using 256-color palette
func Background256(colorNumber int, text string) string {
	if !supports256Color() {
		return text // Fallback to no background
	}
	return fmt.Sprintf("\033[48;5;%dm%s\033[0m", colorNumber, text)
}

// fallbackColor maps 256 colors to nearest standard color
func fallbackColor(colorNumber int, text string) string {
	switch {
	case colorNumber < 8:
		// Standard colors 0-7
		return wrap(fmt.Sprintf("%d", 30+colorNumber), text)
	case colorNumber < 16:
		// Bright colors 8-15
		return wrap(fmt.Sprintf("%d", 82+colorNumber), text)
	case colorNumber >= 232:
		// Grayscale - use white or black
		if colorNumber > 243 {
			return White(text)
		}
		return Black(text)
	default:
		// Color cube - rough approximation
		r := (colorNumber - 16) / 36
		g := ((colorNumber - 16) % 36) / 6
		b := (colorNumber - 16) % 6

		// Convert to nearest standard color
		if r > g && r > b {
			return Red(text)
		} else if g > r && g > b {
			return Green(text)
		} else if b > r && b > g {
			return Blue(text)
		}
		return text // Default
	}
}

// RGB converts RGB values to 256-color palette
func RGB(r, g, b int, text string) string {
	// Convert RGB (0-255) to 256-color palette
	// Formula: 16 + (36 * r/255 * 5) + (6 * g/255 * 5) + (b/255 * 5)
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return text // Invalid RGB values
	}

	colorNumber := 16 + (36 * (r * 5 / 255)) + (6 * (g * 5 / 255)) + (b * 5 / 255)
	return Color256(colorNumber, text)
}

// =============================================================================
// TERMINAL CAPABILITY DETECTION
// =============================================================================

type TerminalInfo struct {
	Name              string
	SupportsColor     bool
	Supports256       bool
	SupportsTrueColor bool
	Width             int
	Height            int
}

// detectTerminalCapabilities checks terminal capabilities
func DetectTerminalCapabilities() TerminalInfo {
	term := os.Getenv("TERM")
	colorterm := os.Getenv("COLORTERM")

	info := TerminalInfo{
		Name: term,
	}

	// Basic color support
	info.SupportsColor = term != "dumb" && term != "" && isTerminal()

	// 256-color support
	info.Supports256 = strings.Contains(term, "256") ||
		strings.Contains(term, "256color") ||
		strings.Contains(term, "xterm")

	// True color support
	info.SupportsTrueColor = strings.Contains(colorterm, "truecolor") ||
		strings.Contains(colorterm, "24bit") ||
		strings.Contains(term, "direct")

	return info
}

// isTerminal checks if output is going to a terminal
func isTerminal() bool {
	// Simple check - in production you'd use golang.org/x/term
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// SafeColor applies color only if terminal supports it
func SafeColor(colorFunc func(string) string, text string) string {
	info := DetectTerminalCapabilities()
	if info.SupportsColor {
		return colorFunc(text)
	}
	return text
}

// =============================================================================
// THEMES (as suggested in homework)
// =============================================================================

type Theme struct {
	Error   func(string) string
	Warning func(string) string
	Success func(string) string
	Info    func(string) string
}

var (
	DefaultTheme = Theme{
		Error:   Red,
		Warning: Yellow,
		Success: Green,
		Info:    Blue,
	}

	DarkTheme = Theme{
		Error:   BoldRed,
		Warning: BoldYellow,
		Success: BoldGreen,
		Info:    Cyan,
	}

	MonochromeTheme = Theme{
		Error:   Bold,
		Warning: Underline,
		Success: func(s string) string { return s },
		Info:    Italic,
	}

	HighContrastTheme = Theme{
		Error:   func(s string) string { return BoldRed(WhiteBg(s)) },
		Warning: func(s string) string { return BoldYellow(BlackBg(s)) },
		Success: func(s string) string { return BoldGreen(BlackBg(s)) },
		Info:    func(s string) string { return BoldBlue(WhiteBg(s)) },
	}
)

var currentTheme = DefaultTheme

// SetTheme sets the global theme
func SetTheme(theme Theme) {
	currentTheme = theme
}

// Error prints error message using current theme
func Error(text string) string {
	return currentTheme.Error(text)
}

// Warning prints warning message using current theme
func Warning(text string) string {
	return currentTheme.Warning(text)
}

// Success prints success message using current theme
func Success(text string) string {
	return currentTheme.Success(text)
}

// Info prints info message using current theme
func Info(text string) string {
	return currentTheme.Info(text)
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// ansiRegex for stripping ANSI codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// StripANSI removes ANSI escape codes from string
func StripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// VisualLength returns the visual length of string (without ANSI codes)
func VisualLength(str string) int {
	return len(StripANSI(str))
}

// ShowColorPalette displays all 256 colors in a grid format
func ShowColorPalette() {
	fmt.Println("=== 256 Color Palette ===")

	// Standard colors (0-15)
	fmt.Println("Standard Colors (0-15):")
	for i := range 16 {
		fmt.Printf("%s ", Color256(i, fmt.Sprintf("%3d", i)))
		if i == 7 || i == 15 {
			fmt.Println()
		}
	}

	// Color cube (16-231)
	fmt.Println("\nColor Cube (16-231):")
	for i := 16; i < 232; i++ {
		fmt.Printf("%s ", Color256(i, "██"))
		if (i-16)%6 == 5 {
			fmt.Print(" ")
		}
		if (i-16)%36 == 35 {
			fmt.Println()
		}
	}

	// Grayscale (232-255)
	fmt.Println("\nGrayscale (232-255):")
	for i := 232; i < 256; i++ {
		fmt.Printf("%s ", Color256(i, "██"))
	}
	fmt.Println()
}

// =============================================================================
// LOG LEVEL IMPLEMENTATION (from transcription examples)
// =============================================================================

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return Cyan("DEBUG")
	case INFO:
		return Blue("INFO")
	case WARN:
		return Yellow("WARN")
	case ERROR:
		return Red("ERROR")
	default:
		return "UNKNOWN"
	}
}

// ColoredProgressBar creates a colored progress bar
func ColoredProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	empty := width - filled

	var bar strings.Builder

	// Green for completed portion
	bar.WriteString(Green(strings.Repeat("█", filled)))

	// Gray for empty portion
	bar.WriteString(Cyan(strings.Repeat("░", empty)))

	percentage := int(progress * 100)
	return fmt.Sprintf("%s %3d%%", bar.String(), percentage)
}

// ShowStatus displays a status indicator with color
func ShowStatus(name string, success bool) string {
	status := "✗"
	colorFunc := Red

	if success {
		status = "✓"
		colorFunc = Green
	}

	return fmt.Sprintf("%s %s", colorFunc(status), name)
}

// =============================================================================
// ENVIRONMENT VARIABLE SUPPORT
// =============================================================================

// respects NO_COLOR environment variable
func isColorDisabled() bool {
	return os.Getenv("NO_COLOR") != ""
}

// ConditionalColor applies color only if not disabled by environment
func ConditionalColor(colorFunc func(string) string, text string) string {
	if isColorDisabled() || !isTerminal() {
		return text
	}
	return colorFunc(text)
}
