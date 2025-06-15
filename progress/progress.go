// Package progress provides a customizable terminal progress bar
package progress

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

// BarConfig holds configuration options for the progress bar
type BarConfig struct {
	Width       int       // Fixed width (0 = auto-detect terminal width)
	FilledChar  string    // Character for filled portion
	EmptyChar   string    // Character for empty portion
	Writer      io.Writer // Output destination
	ShowPercent bool      // Whether to show percentage
	ShowETA     bool      // Whether to show estimated time remaining
}

// Bar represents a terminal progress bar
type Bar struct {
	config       BarConfig
	totalWidth   int
	lastProgress float64
	started      bool
	stopped      bool
	startTime    time.Time
	termSizeCh   chan os.Signal
	lock         sync.RWMutex
}

// Option represents a configuration option for the progress bar
type Option func(*BarConfig)

// WithWidth sets a fixed width for the progress bar
func WithWidth(width int) Option {
	return func(c *BarConfig) {
		c.Width = width
	}
}

// WithFilledChar sets the character used for the filled portion
func WithFilledChar(char string) Option {
	return func(c *BarConfig) {
		c.FilledChar = char
	}
}

// WithEmptyChar sets the character used for the empty portion
func WithEmptyChar(char string) Option {
	return func(c *BarConfig) {
		c.EmptyChar = char
	}
}

// WithWriter sets the output destination
func WithWriter(writer io.Writer) Option {
	return func(c *BarConfig) {
		c.Writer = writer
	}
}

// WithPercent enables/disables percentage display
func WithPercent(show bool) Option {
	return func(c *BarConfig) {
		c.ShowPercent = show
	}
}

// WithETA enables/disables estimated time remaining
func WithETA(show bool) Option {
	return func(c *BarConfig) {
		c.ShowETA = show
	}
}

// Predefined styles
var (
	StyleDefault = BarConfig{
		FilledChar:  "#",
		EmptyChar:   " ",
		Writer:      os.Stdout,
		ShowPercent: true,
		ShowETA:     false,
	}

	StyleBlocks = BarConfig{
		FilledChar:  "█",
		EmptyChar:   "░",
		Writer:      os.Stdout,
		ShowPercent: true,
		ShowETA:     false,
	}

	StyleDots = BarConfig{
		FilledChar:  "●",
		EmptyChar:   "○",
		Writer:      os.Stdout,
		ShowPercent: true,
		ShowETA:     false,
	}

	StyleMinimal = BarConfig{
		FilledChar:  "=",
		EmptyChar:   "-",
		Writer:      os.Stdout,
		ShowPercent: false,
		ShowETA:     false,
	}
)

// NewBar creates a new progress bar with default configuration
func NewBar() *Bar {
	return NewBarWithConfig(StyleDefault)
}

// NewBarWithStyle creates a new progress bar with a predefined style
func NewBarWithStyle(style BarConfig) *Bar {
	return NewBarWithConfig(style)
}

// NewBarWithConfig creates a new progress bar with custom configuration
func NewBarWithConfig(config BarConfig, opts ...Option) *Bar {
	// Apply options to config
	for _, opt := range opts {
		opt(&config)
	}

	// Set defaults if not specified
	if config.FilledChar == "" {
		config.FilledChar = "#"
	}
	if config.EmptyChar == "" {
		config.EmptyChar = " "
	}
	if config.Writer == nil {
		config.Writer = os.Stdout
	}

	b := &Bar{
		config:       config,
		lastProgress: 0,
		termSizeCh:   make(chan os.Signal, 1),
	}

	b.calculateWidth()
	return b
}

// calculateWidth determines the width of the progress bar
func (b *Bar) calculateWidth() {
	if b.config.Width > 0 {
		b.totalWidth = b.config.Width
		return
	}

	// Auto-detect terminal width
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		b.totalWidth = 60 // Fallback width
		return
	}

	// Reserve space for percentage and ETA
	reservedSpace := 0
	if b.config.ShowPercent {
		reservedSpace += 5 // " 100%"
	}
	if b.config.ShowETA {
		reservedSpace += 12 // " ETA: 00:00"
	}

	b.totalWidth = width - reservedSpace - 2 // -2 for brackets or margins
	if b.totalWidth < 10 {
		b.totalWidth = 10 // Minimum width
	}
}

// clearLine clears the current terminal line
func (b *Bar) clearLine() {
	totalClearWidth := b.totalWidth
	if b.config.ShowPercent {
		totalClearWidth += 5
	}
	if b.config.ShowETA {
		totalClearWidth += 12
	}

	empty := strings.Repeat(" ", totalClearWidth+5) // +5 for safety margin
	fmt.Fprintf(b.config.Writer, "\r%s", empty)
}

// Start initializes the progress bar
func (b *Bar) Start() {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.started && !b.stopped {
		return // Already started
	}

	b.started = true
	b.stopped = false
	b.startTime = time.Now()
	b.clearLine()
	b.lastProgress = 0

	// Set up terminal resize handling if using auto-width
	if b.config.Width == 0 {
		signal.Notify(b.termSizeCh, syscall.SIGWINCH)
		go b.handleResize()
	}
}

// handleResize manages terminal window resize events
func (b *Bar) handleResize() {
	for {
		select {
		case <-b.termSizeCh:
			b.lock.Lock()
			if b.stopped {
				b.lock.Unlock()
				return
			}
			b.calculateWidth()
			progress := b.lastProgress
			b.lock.Unlock()
			b.SetProgress(progress) // Redraw with new width
		}
	}
}

// Stop cleans up the progress bar
func (b *Bar) Stop() {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.stopped {
		return
	}

	b.stopped = true

	// Clean up signal handling
	if b.config.Width == 0 {
		signal.Stop(b.termSizeCh)
	}

	b.clearLine()
	fmt.Fprintf(b.config.Writer, "\r")
}

// SetProgress updates the progress (0.0 to 1.0)
func (b *Bar) SetProgress(progress float64) {
	b.lock.RLock()
	if !b.started || b.stopped {
		b.lock.RUnlock()
		return
	}
	b.lock.RUnlock()

	// Constrain progress to valid range
	progress = math.Max(0.0, math.Min(1.0, progress))

	b.lock.Lock()
	b.lastProgress = progress
	b.lock.Unlock()

	// Calculate filled and empty portions
	filledCount := int(float64(b.totalWidth) * progress)
	emptyCount := b.totalWidth - filledCount

	// Build progress bar string
	var bar strings.Builder
	bar.WriteString("\r")

	// Write filled portion
	for i := 0; i < filledCount; i++ {
		bar.WriteString(b.config.FilledChar)
	}

	// Write empty portion
	for i := 0; i < emptyCount; i++ {
		bar.WriteString(b.config.EmptyChar)
	}

	// Add percentage if enabled
	if b.config.ShowPercent {
		percentage := int(progress * 100)
		bar.WriteString(fmt.Sprintf(" %3d%%", percentage))
	}

	// Add ETA if enabled
	if b.config.ShowETA {
		eta := b.calculateETA(progress)
		bar.WriteString(fmt.Sprintf(" ETA: %s", eta))
	}

	fmt.Fprint(b.config.Writer, bar.String())
}

// calculateETA estimates time remaining based on current progress
func (b *Bar) calculateETA(progress float64) string {
	if progress <= 0 {
		return "--:--"
	}

	elapsed := time.Since(b.startTime)
	totalEstimated := time.Duration(float64(elapsed) / progress)
	remaining := totalEstimated - elapsed

	if remaining < 0 {
		remaining = 0
	}

	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// Run executes a function with progress updates
func (b *Bar) Run(fn func(setProgress func(float64))) {
	b.Start()
	defer b.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		fn(b.SetProgress)
	}()

	wg.Wait()
}

// Increment increases progress by a delta amount
func (b *Bar) Increment(delta float64) {
	b.lock.RLock()
	current := b.lastProgress
	b.lock.RUnlock()

	b.SetProgress(current + delta)
}

// GetProgress returns the current progress value
func (b *Bar) GetProgress() float64 {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.lastProgress
}

// IsStarted returns whether the progress bar has been started
func (b *Bar) IsStarted() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.started
}

// IsStopped returns whether the progress bar has been stopped
func (b *Bar) IsStopped() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.stopped
}

// Reset resets the progress bar to initial state
func (b *Bar) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.lastProgress = 0
	b.started = false
	b.stopped = false
	b.startTime = time.Time{}
}

// MultiBar manages multiple progress bars
type MultiBar struct {
	bars map[string]*LabeledBar
	lock sync.RWMutex
}

// LabeledBar represents a progress bar with a label
type LabeledBar struct {
	*Bar
	label string
	line  int
}

// NewMultiBar creates a new multi-bar manager
func NewMultiBar() *MultiBar {
	return &MultiBar{
		bars: make(map[string]*LabeledBar),
	}
}

// Add adds a labeled progress bar
func (mb *MultiBar) Add(name, label string, config BarConfig, opts ...Option) {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	bar := NewBarWithConfig(config, opts...)
	mb.bars[name] = &LabeledBar{
		Bar:   bar,
		label: label,
		line:  len(mb.bars),
	}
}

// Start starts a specific progress bar
func (mb *MultiBar) Start(name string) {
	mb.lock.RLock()
	defer mb.lock.RUnlock()

	if bar, exists := mb.bars[name]; exists {
		fmt.Printf("%s:\n", bar.label)
		bar.Start()
	}
}

// Stop stops a specific progress bar
func (mb *MultiBar) Stop(name string) {
	mb.lock.RLock()
	defer mb.lock.RUnlock()

	if bar, exists := mb.bars[name]; exists {
		bar.Stop()
		fmt.Printf("%s: Complete!\n", bar.label)
	}
}

// SetProgress sets progress for a specific bar
func (mb *MultiBar) SetProgress(name string, progress float64) {
	mb.lock.RLock()
	defer mb.lock.RUnlock()

	if bar, exists := mb.bars[name]; exists {
		bar.SetProgress(progress)
	}
}

// StopAll stops all progress bars
func (mb *MultiBar) StopAll() {
	mb.lock.RLock()
	defer mb.lock.RUnlock()

	for _, bar := range mb.bars {
		bar.Stop()
	}
}

// Convenience functions

// Quick creates and runs a simple progress bar
func Quick(steps int, fn func(step func())) {
	bar := NewBar()
	bar.Start()
	defer bar.Stop()

	currentStep := 0
	stepFunc := func() {
		currentStep++
		progress := float64(currentStep) / float64(steps)
		bar.SetProgress(progress)
	}

	fn(stepFunc)
}

// WithStyle creates a progress bar with a predefined style
func WithStyle(style BarConfig) *Bar {
	return NewBarWithStyle(style)
}

// DownloadProgress simulates a download progress bar
func DownloadProgress(totalBytes int64, fn func(downloaded func(int64))) {
	bar := NewBarWithConfig(StyleDefault, WithETA(true))
	bar.Start()
	defer bar.Stop()

	downloadedBytes := int64(0)
	downloadFunc := func(bytes int64) {
		downloadedBytes += bytes
		progress := float64(downloadedBytes) / float64(totalBytes)
		bar.SetProgress(progress)
	}

	fn(downloadFunc)
}

