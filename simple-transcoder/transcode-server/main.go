package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"encoding/json"
	"strconv"
	"math"
	"time"
	// "os"

	// "github.com/anacrolix/torrent"
)

type Format struct {
	Duration int `json:"duration"`
}
type Metadata struct {
    Format Format `json:"format"`
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= h * time.Minute
	s := m / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.URL.Query()["metadata"]
    if ok {
		// TODO: get real metadata
		metadata := Metadata {
			Format: Format{
				Duration: 633,
			},
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			panic(err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	ffmpegArgs := []string{"-i", "http://localhost:3000", "-f", "matroska", "-c:v", "libx264", "-b", "300k", "-preset", "fast", "-tune", "zerolatency"}

	timeArgs, ok := r.URL.Query()["time"]
    if ok {
		var timeFloat float64
		timeFloat, err := strconv.ParseFloat(timeArgs[0], 64)
		if err != nil {
			timeFloat = 0
		}
		var timeInt int = int(math.Round(timeFloat))
		timeDuration := time.Second * time.Duration(timeInt)
		timeString := fmtDuration(timeDuration) 
		ffmpegArgs = append(ffmpegArgs, "-ss", timeString)
	}
	ffmpegArgs = append(ffmpegArgs, "-")
	
	cmdFF := exec.Command("ffmpeg", ffmpegArgs...)
	defer func() {
		// TODO: Why doesn't it ever close?
		println("Closing ffmpeg...")
		cmdFF.Process.Kill()
	}()

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
