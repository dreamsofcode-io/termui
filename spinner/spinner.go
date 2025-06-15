// Package spinner provides a customizable terminal loading spinner
package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Frames represents a sequence of animation frames
type Frames = []rune

// Predefined animation frame sets
var (
	// FramesLines - Classic rotating bar animation
	FramesLines = []rune{'|', '/', '-', '\\'}

	// FramesDots - Modern Unicode dot animation
	FramesDots = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

	// FramesBounce - Simple bouncing dots
	FramesBounce = []rune{'.', 'o', 'O', 'o'}

	// FramesArrows - Rotating arrows
	FramesArrows = []rune{'↖', '↗', '↘', '↙'}

	// FramesProgress - Progress-style animation
	FramesProgress = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
)

// Spinner represents a terminal loading spinner
type Spinner struct {
	frames        Frames
	frameDuration time.Duration
	writer        io.Writer
	prefix        string
	suffix        string
	doneCh        chan struct{}
	finishedCh    chan struct{}
	lock          sync.Mutex
	running       bool
}

// Option represents a configuration option for the spinner
type Option func(*Spinner)

// WithFrameDuration sets the duration between animation frames
func WithFrameDuration(duration time.Duration) Option {
	return func(s *Spinner) {
		s.frameDuration = duration
	}
}

// WithFrames sets the animation frames to use
func WithFrames(frames Frames) Option {
	return func(s *Spinner) {
		s.frames = frames
	}
}

// WithWriter sets the output writer (defaults to os.Stdout)
func WithWriter(writer io.Writer) Option {
	return func(s *Spinner) {
		s.writer = writer
	}
}

// WithPrefix sets text to display before the spinner
func WithPrefix(prefix string) Option {
	return func(s *Spinner) {
		s.prefix = prefix
	}
}

// WithSuffix sets text to display after the spinner
func WithSuffix(suffix string) Option {
	return func(s *Spinner) {
		s.suffix = suffix
	}
}

// New creates a new spinner with the given options
func New(opts ...Option) *Spinner {
	s := &Spinner{
		frames:        FramesLines,
		frameDuration: 100 * time.Millisecond,
		writer:        os.Stdout,
		prefix:        "",
		suffix:        "",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Prevent multiple starts
	if s.running {
		return
	}

	doneCh := make(chan struct{})
	finishedCh := make(chan struct{})

	go func() {
		defer close(finishedCh)
		defer s.clearLine()

		ticker := time.NewTicker(s.frameDuration)
		defer ticker.Stop()

		frameIndex := 0

		for {
			select {
			case <-ticker.C:
				frame := s.frames[frameIndex%len(s.frames)]
				fmt.Fprintf(s.writer, "\r%s%c%s", s.prefix, frame, s.suffix)
				frameIndex++

			case <-doneCh:
				return
			}
		}
	}()

	s.doneCh = doneCh
	s.finishedCh = finishedCh
	s.running = true
}

// Stop stops the spinner animation and cleans up
func (s *Spinner) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.running || s.doneCh == nil {
		return
	}

	close(s.doneCh)
	<-s.finishedCh

	s.doneCh = nil
	s.finishedCh = nil
	s.running = false
}

// IsRunning returns whether the spinner is currently running
func (s *Spinner) IsRunning() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.running
}

// SetPrefix updates the prefix text (can be called while running)
func (s *Spinner) SetPrefix(prefix string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.prefix = prefix
}

// SetSuffix updates the suffix text (can be called while running)
func (s *Spinner) SetSuffix(suffix string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.suffix = suffix
}

// clearLine clears the current line in the terminal
func (s *Spinner) clearLine() {
	// Calculate the total width to clear
	maxWidth := len(s.prefix) + len(s.suffix) + 1 // +1 for spinner character
	clearStr := make([]byte, maxWidth)
	for i := range clearStr {
		clearStr[i] = ' '
	}
	fmt.Fprintf(s.writer, "\r%s\r", clearStr)
}

// Restart stops and then starts the spinner (useful for changing options)
func (s *Spinner) Restart() {
	s.Stop()
	s.Start()
}

// Run runs a function while displaying the spinner
func (s *Spinner) Run(fn func()) {
	s.Start()
	defer s.Stop()
	fn()
}

// RunWithTimeout runs a function with a spinner and timeout
func (s *Spinner) RunWithTimeout(fn func() error, timeout time.Duration) error {
	s.Start()
	defer s.Stop()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// MultiSpinner manages multiple labeled spinners
type MultiSpinner struct {
	spinners map[string]*LabeledSpinner
	lock     sync.RWMutex
}

// LabeledSpinner represents a spinner with a label
type LabeledSpinner struct {
	*Spinner
	label string
	line  int
}

// NewMultiSpinner creates a new multi-spinner manager
func NewMultiSpinner() *MultiSpinner {
	return &MultiSpinner{
		spinners: make(map[string]*LabeledSpinner),
	}
}

// Add adds a labeled spinner to the multi-spinner
func (ms *MultiSpinner) Add(name, label string, opts ...Option) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	// Set prefix to include label
	opts = append(opts, WithPrefix(label+" "))

	spinner := New(opts...)
	ms.spinners[name] = &LabeledSpinner{
		Spinner: spinner,
		label:   label,
		line:    len(ms.spinners),
	}
}

// Start starts a specific spinner by name
func (ms *MultiSpinner) Start(name string) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	if spinner, exists := ms.spinners[name]; exists {
		spinner.Start()
	}
}

// Stop stops a specific spinner by name
func (ms *MultiSpinner) Stop(name string) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	if spinner, exists := ms.spinners[name]; exists {
		spinner.Stop()
	}
}

// StartAll starts all spinners
func (ms *MultiSpinner) StartAll() {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	for _, spinner := range ms.spinners {
		spinner.Start()
	}
}

// StopAll stops all spinners
func (ms *MultiSpinner) StopAll() {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	for _, spinner := range ms.spinners {
		spinner.Stop()
	}
}

// UpdateLabel updates the label for a specific spinner
func (ms *MultiSpinner) UpdateLabel(name, newLabel string) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	if spinner, exists := ms.spinners[name]; exists {
		spinner.SetPrefix(newLabel + " ")
		spinner.label = newLabel
	}
}

// Convenience functions for common use cases

// WithMessage creates a spinner with a prefix message
func WithMessage(message string) *Spinner {
	return New(WithPrefix(message + " "))
}

// Quick starts a spinner with a message and returns stop function
func Quick(message string) func() {
	s := WithMessage(message)
	s.Start()
	return s.Stop
}

// Perform runs a function with a spinner and message
func Perform(message string, fn func()) {
	s := WithMessage(message)
	s.Run(fn)
}

