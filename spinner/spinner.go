package spinner

import (
	"fmt"
	"sync"
	"time"
)

type Frames = []rune

var (
	FramesLines = []rune{'|', '/', '-', '\\'}
	FramesDots  = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}
)

type Spinner struct {
	frames        Frames
	frameDuration time.Duration
	doneCh        chan struct{}
	finishedCh    chan struct{}
	lock          sync.Mutex
}

type Option func(*Spinner)

func WithFrameDuration(duration time.Duration) Option {
	return func(s *Spinner) {
		s.frameDuration = duration
	}
}

func WithFrames(frames Frames) Option {
	return func(s *Spinner) {
		s.frames = frames
	}
}

func New(opts ...Option) *Spinner {
	s := &Spinner{
		frames:        FramesLines,
		frameDuration: time.Millisecond * 100,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Spinner) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.doneCh != nil {
		return
	}

	doneCh := make(chan struct{})
	closedCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(s.frameDuration)
		i := 0

		for {
			i++
			select {
			case <-ticker.C:
				fmt.Printf("\r%c", s.frames[i%len(s.frames)])
			case <-doneCh:
				fmt.Print("\r")
				close(closedCh)
				return
			}
		}
	}()

	s.doneCh = doneCh
	s.finishedCh = closedCh
}

func (s *Spinner) Stop() {
	if s.doneCh == nil {
		return
	}

	close(s.doneCh)
	<-s.finishedCh
}
