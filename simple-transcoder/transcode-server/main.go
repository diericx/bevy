package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	// "time"
	// "os"

	// "github.com/anacrolix/torrent"
)

func fileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Starting transcode...")

	ffmpegArgs := []string{"-i", "http://localhost:3000", "-f", "matroska", "-c:v", "libx264", "-b", "300k", "-preset", "fast", "-tune", "zerolatency"}
	time, ok := r.URL.Query()["time"]
    if ok {
		println(time)
		ffmpegArgs = append(ffmpegArgs, "-ss", time[0])
	}
	ffmpegArgs = append(ffmpegArgs, "-")
	
	fmt.Println(ffmpegArgs)

	cmdFF := exec.Command("ffmpeg", ffmpegArgs...)
	defer cmdFF.Process.Kill()

	cmdFF.Stdout = w
    if err := cmdFF.Run(); err != nil {
        log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", fileHandler)

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
