package progress

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/term"
)

type Bar struct {
	totalWidth   int
	lastPosition float64
	termSizeCh   chan os.Signal
}

func NewBar() *Bar {
	b := &Bar{
		lastPosition: 0,
		termSizeCh:   make(chan os.Signal),
	}

	b.calculateWidth()

	return b
}

func (b *Bar) calculateWidth() {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	b.totalWidth = width - 1
}

func (b *Bar) clearLine() {
	empty := strings.Repeat(" ", b.totalWidth)
	fmt.Printf("\r%s", empty)
}

func (b *Bar) Run(f func(setProgress func(progress float64))) {
	b.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		f(b.SetProgress)
	}()

	b.Stop()
}

func (b *Bar) Start() {
	b.clearLine()
	b.lastPosition = 0

	signal.Notify(b.termSizeCh, syscall.SIGWINCH)

	go func() {
		<-b.termSizeCh
		width, _, _ := term.GetSize(int(os.Stdout.Fd()))
		b.totalWidth = width - 1
		b.SetProgress(b.lastPosition)
	}()
}

func (b *Bar) Stop() {
	b.clearLine()
	fmt.Print("\r")
}

func (b *Bar) SetProgress(progress float64) {
	b.lastPosition = progress

	filledCount := int(float64(b.totalWidth) * progress)
	emptyCount := b.totalWidth - filledCount

	filled := strings.Repeat("#", filledCount)
	empty := strings.Repeat(" ", emptyCount)

	fmt.Printf("\r%s%s", filled, empty)
}
