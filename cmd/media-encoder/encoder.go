package main

import (
	"bytes"
	"fmt"
	"os/exec"
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

func main() {
	dur, _, err := Shellout(`ffprobe -i ./media/neuro-video-30.mp4 -show_entries format=duration -v quiet -of csv="p=0"`)
	if err != nil {
		panic(fmt.Sprintf("Error running command: %v", err))
	}

	// command to split the videos
	_, _, err = Shellout("ffmpeg -ss 00:00:01 -i ./media/neuro-video-30.mp4 -t 00:00:30 -c copy ./split/split.mp4")
	if err != nil {
		panic(fmt.Sprintf("Error running command: %v", err))
	}

}
