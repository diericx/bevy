package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/anacrolix/torrent"
)

var cl *torrent.Client

func torHandler(w http.ResponseWriter, r *http.Request) {
	// add the magnet (in a round about way so we can log if it was already seen)
	// uri := "magnet:?xt=urn:btih:88594AAACBDE40EF3E2510C47374EC0AA396C08E&dn=bbb_sunflower_1080p_30fps_normal.mp4&tr=udp%3a%2f%2ftracker.openbittorrent.com%3a80%2fannounce&tr=udp%3a%2f%2ftracker.publicbt.com%3a80%2fannounce&ws=http%3a%2f%2fdistribution.bbb3d.renderfarming.net%2fvideo%2fmp4%2fbbb_sunflower_1080p_30fps_normal.mp4"
	// t, err := cl.AddMagnet(uri)

	t, err := cl.AddTorrentFromFile("Ultimate.Kids.Movies.Pack.HD.V2-PULSAR.torrent")
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "couldnt add the magnet")
		log.Printf("ERROR: %v\n", err)
		return
	}

	// wait for info
	<-t.GotInfo()
	name := t.Name()
	files := t.Files()
	log.Println("***FILES:", files)
	log.Println("***FILE:", files[0].FileInfo())
	log.Println("***Hash:", t.Metainfo().HashInfoBytes().AsString())
	info, _ := t.Metainfo().UnmarshalInfo()
	log.Println("***Info:", info.Files)
	log.Println("***Name:", t.Name())

	// mark the whole thing for download but prio the treader?
	treader := files[0].NewReader()
	defer treader.Close()

	http.ServeContent(w, r, name, time.Time{}, treader)
}

func main() {
	client, err := torrent.NewClient(nil)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
	}
	defer client.Close()
	cl = client

	http.HandleFunc("/", torHandler)

	log.Println("Listening on :3000...")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
