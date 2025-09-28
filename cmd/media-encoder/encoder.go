package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const ShellToUse = "bash"

func Shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// var VIDEO_PATH = "./media/neuro-less.mp4"
var VIDEO_PATH = "./media/neuro-30min.mp4"
var OUTPUT_PATH = "./media/split"
var SPLIT_TIME = time.Minute
var MAX_GOROUTINES = 10

type Timespan time.Duration

func (t Timespan) Format(format string) string {
	return time.Unix(0, 0).UTC().Add(time.Duration(t)).Format(format)
}

func SplitVideo(processId int, buf chan struct{}, wg *sync.WaitGroup) {
	_, _, err := Shellout(fmt.Sprintf(
		"ffmpeg -ss %s -i %s -t %s -c copy %s/split_%d.mp4 -y",
		Timespan(SPLIT_TIME*time.Duration(processId)).Format("15:04:05"),
		VIDEO_PATH,
		Timespan(SPLIT_TIME).Format("15:04:05"),
		OUTPUT_PATH,
		processId,
	))
	if err != nil {
		panic(fmt.Sprintf("Error running command: %v", err))
	}

	<-buf
	wg.Done()
}

func main() {
	err := os.MkdirAll(OUTPUT_PATH, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Error making directory: %v", err))
	}

	var ch = make(chan struct{}, MAX_GOROUTINES)
	defer close(ch)
	var wg sync.WaitGroup

	dur, _, err := Shellout(fmt.Sprintf(`ffprobe -i %s -show_entries format=duration -v quiet -of csv="p=0"`, VIDEO_PATH))
	if err != nil {
		panic(fmt.Sprintf("Error running command: %v", err))
	}

	dur = strings.ReplaceAll(dur, "\x0a", "")

	vidDur, err := time.ParseDuration(dur + "s")
	if err != nil {
		panic(fmt.Sprintf("Error parsing video duration: %v", err))
	}

	parts := int(math.Ceil(vidDur.Minutes()))

	// probably it will need to store some data about
	// the video itself, like how many chunks it have
	// which is the first and some other stuff to make it
	// possible to find it i guess

	processId := 0
	done := false
	for {
		ch <- struct{}{}

		wg.Add(1)
		go SplitVideo(processId, ch, &wg)
		processId++

		if parts == processId {
			done = true
		}

		if done {
			break
		}
	}

	wg.Wait()
	fmt.Println("Waiting")
}
