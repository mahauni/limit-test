package main

import (
	"bytes"
	"fmt"
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

var VIDEO_PATH = "./media/neuro-30-min.mp4"
var OUTPUT_PATH = "./media/split"
var SPLIT_TIME = time.Minute
var MAX_GOROUTINES = 10

type Timespan time.Duration

func (t Timespan) Format(format string) string {
	return time.Unix(0, 0).UTC().Add(time.Duration(t)).Format(format)
}

func SplitVideo(processId int, buf chan struct{}, output chan int, wg *sync.WaitGroup) {
	defer wg.Done()
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

	fmt.Println(processId)
	<-buf
	output <- processId
}

func main() {
	var ch = make(chan struct{}, MAX_GOROUTINES)
	defer close(ch)
	var output = make(chan int)
	defer close(output)
	var wg sync.WaitGroup

	dur, _, err := Shellout(fmt.Sprintf(`ffprobe -i %s -show_entries format=duration -v quiet -of csv="p=0"`, VIDEO_PATH))
	if err != nil {
		panic(fmt.Sprintf("Error running command: %v", err))
	}

	dur = strings.ReplaceAll(dur, "\u00a0", "")

	vidDur, err := time.ParseDuration("1609.352000s")
	if err != nil {
		panic(fmt.Sprintf("Error parsing video duration: %v", err))
	}

	processId := 0
	done := false
	for {
		ch <- struct{}{}

		select {
		case x := <-output:
			// if SPLIT_TIME*time.Duration(x) < vidDur {
			if x+MAX_GOROUTINES < 26 {

				fmt.Println(SPLIT_TIME*time.Duration(x), vidDur)
				wg.Add(1)
				go SplitVideo(x+MAX_GOROUTINES, ch, output, &wg)
			} else {
				// this is not the best one because it still
				// probably will skip some routines if it is faster
				done = true
			}
		default:
			wg.Add(1)
			go SplitVideo(processId, ch, output, &wg)
		}

		if done {
			break
		}

		if processId < MAX_GOROUTINES {
			processId++
		}
	}

	wg.Wait()
	fmt.Println("Waiting")
}
