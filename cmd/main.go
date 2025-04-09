package main

import (
	"fmt"
	"time"

	"github.com/dreamsofcode-io/ui/spinner"
)

func main() {
	fmt.Println("Starting...")

	s := spinner.New(spinner.WithFrames(spinner.FramesDots))
	s.Start()

	time.Sleep(time.Second * 3)

	s.Stop()

	fmt.Println("Finished!")
}
