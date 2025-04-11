package main

import (
	"fmt"
	"time"

	"github.com/dreamsofcode-io/termui/progress"
)

func main() {
	bar := progress.NewBar()

	fmt.Println("Starting...")

	bar.Start()

	const steps = 10
	for i := range steps {
		progress := 1.0 / float64(steps) * float64(i+1)
		bar.SetProgress(progress)
		time.Sleep(time.Millisecond * 500)
	}

	bar.Stop()

	fmt.Println("Finished!")
}
