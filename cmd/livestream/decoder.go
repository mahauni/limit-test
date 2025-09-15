package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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
var OUTPUT_PATH = "./media/intermediate" // review this
var INPUT_FILE_NAME = "split"
var NUM_CHUNKS = 27
var MAX_GOROUTINES = 10

var SPLIT_TIME = time.Minute

type Timespan time.Duration

func (t Timespan) Format(format string) string {
	return time.Unix(0, 0).UTC().Add(time.Duration(t)).Format(format)
}

func SplitVideo(processId int, buf chan struct{}, output chan string, wg *sync.WaitGroup) {
	file_output := fmt.Sprintf("%s/intermediate_%d.ts", OUTPUT_PATH, processId)
	_, _, err := Shellout(fmt.Sprintf(
		"ffmpeg -i %s -c copy -bsf:v h264_mp4toannexb -f mpegts %s -y",
		fmt.Sprintf("%s/%s_%d.mp4", VIDEO_PATH, INPUT_FILE_NAME, processId),
		file_output,
	))
	if err != nil {
		panic(fmt.Sprintf("Error running command: %v", err))
	}

	<-buf
	wg.Done()
	output <- file_output
}

func main() {
	err := os.MkdirAll(OUTPUT_PATH, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Error making directory: %v", err))
	}

	var ch = make(chan struct{}, MAX_GOROUTINES)
	defer close(ch)
	var output = make(chan string)
	defer close(output)
	var wg sync.WaitGroup

	partsStr, _, err := Shellout(fmt.Sprintf(
		`ls %s | grep "%s" | wc -l`,
		VIDEO_PATH,
		INPUT_FILE_NAME,
	))
	if err != nil {
		panic(fmt.Sprintf("Error running command: %v", err))
	}
	parts, err := strconv.Atoi(strings.TrimSuffix(partsStr, "\n"))
	if err != nil {
		panic(fmt.Sprintf("Error convering to int: %v", err))
	}

	processId := 0
	done := false
	for {
		ch <- struct{}{}

		wg.Add(1)
		go SplitVideo(processId, ch, output, &wg)
		processId++

		if parts == processId {
			done = true
		}

		if done {
			break
		}
	}

	wg.Wait()

	m3u8 := ""

	m3u8 += fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:%.6f\n#EXT-X-MEDIA-SEQUENCE:0\n", (SPLIT_TIME + time.Second).Seconds())

	for i := range parts {
		dur, _, err := Shellout(fmt.Sprintf(
			"ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 %s",
			fmt.Sprintf("%s/intermediate_%d.ts", OUTPUT_PATH, i),
		))
		if err != nil {
			panic(fmt.Sprintf("Error running command: %v", err))
		}
		m3u8 += fmt.Sprintf("#EXTINF:%s,\nintermediate_%d.ts\n", strings.TrimSuffix(dur, "\n"), i)
	}

	m3u8 += "#EXT-X-ENDLIST\n"

	err = os.WriteFile(fmt.Sprintf("%s/output.m3u8", OUTPUT_PATH), []byte(m3u8), 0644)
	if err != nil {
		panic(fmt.Sprintf("Error writing to file: %v", err))
	}
}
