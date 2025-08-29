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

// These are the commands that are needed to execute....
// there is 1 command for each chunk
// plus 2 more to get the m3u8

// ffmpeg -i input1.mp4 -c copy -bsf:v h264_mp4toannexb -f mpegts intermediate1.ts
// ffmpeg -i input2.mp4 -c copy -bsf:v h264_mp4toannexb -f mpegts intermediate2.ts
// ffmpeg -i input3.mp4 -c copy -bsf:v h264_mp4toannexb -f mpegts intermediate3.ts
// ffmpeg -i "concat:intermediate1.ts|intermediate2.ts|intermediate3.ts" -c copy -bsf:a aac_adtstoasc -f mpegts intermediate_all.ts
// ffmpeg -i intermediate_all.ts -c copy -bsf:a aac_adtstoasc -f hls -hls_time 10 -hls_list_size 0 output.m3u8

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

var VIDEO_PATH = "./media/split"
var OUTPUT_PATH = "./media" // review this
var NUM_CHUNKS = 27
var MAX_GOROUTINES = 10

var SPLIT_TIME = time.Minute

type Timespan time.Duration

func (t Timespan) Format(format string) string {
	return time.Unix(0, 0).UTC().Add(time.Duration(t)).Format(format)
}

func SplitVideo(processId int, buf chan struct{}, output chan int, wg *sync.WaitGroup) {
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
	output <- processId
}

func main() {
	err := os.MkdirAll(OUTPUT_PATH, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Error making directory: %v", err))
	}

	var ch = make(chan struct{}, MAX_GOROUTINES)
	defer close(ch)
	var output = make(chan int)
	defer close(output)
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

		select {
		case x := <-output:
			if SPLIT_TIME*time.Duration(x+MAX_GOROUTINES) < vidDur {
				wg.Add(1)
				parts--
				go SplitVideo(x+MAX_GOROUTINES, ch, output, &wg)
			}

			if parts == 0 {
				done = true
			}
		default:
			if processId < MAX_GOROUTINES {
				wg.Add(1)
				parts--
				go SplitVideo(processId, ch, output, &wg)
			}

			if parts == 0 {
				done = true
			}
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
