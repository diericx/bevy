package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"os"

	// "github.com/anacrolix/torrent"
)

func fileHandler(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "video/x-flv")
	// w.Header().Set("Transfer-Encoding", "chunked")
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Content-Length", "276134947")
	// w.WriteHeader(200)

	f, err := os.Open("bbb_sunflower_1080p_30fps_normal.mp4")
	if err != nil {
		fmt.Println(err)
	}

	http.ServeContent(w, r, "bbb_sunflower_1080p_30fps_normal.mp4", time.Time{}, f)
}

func main() {
	http.HandleFunc("/", fileHandler)

	log.Println("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
