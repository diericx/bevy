package ffmpeg

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/diericx/iceetime/internal/app"
)

// NewTranscodeCommand returns a new command to transcode media with given constraints
func NewTranscodeCommand(input string, time string, resolution string, maxBitrate string, c app.FFMPEGConfig, audioStream int, videoStream int) *exec.Cmd {
	// Note: -ss flag needs to come before -i in order to skip encoding the entire first section
	ffmpegArgs := []string{
		"-ss", time,
		"-i", input,
		"-f", c.Video.Format,
		"-c:v", c.Video.CompressionAlgo,
		"-maxrate", maxBitrate,
		"-vf", fmt.Sprintf("scale=%s", resolution),
		"-threads", "0",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-map", fmt.Sprintf("0:v:%v", videoStream),
		"-map", fmt.Sprintf("0:a:%v", audioStream),
		// "-movflags", "frag_keyframe+empty_moov", // This was to allow mp4 encoding.. not sure what it implies
	}

	log.Printf("%+v", ffmpegArgs)

	ffmpegArgs = append(ffmpegArgs, "-")

	cmdFF := exec.Command("ffmpeg", ffmpegArgs...)
	return cmdFF
}
