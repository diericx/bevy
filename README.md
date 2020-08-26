# Iceetime

Iceetime is a self hosted alternative to Popcorntime that aims to improve file availability and increase control over the files being downloaded. 

Torrents are grabbed from indexers by hitting Jackett (no need to reinvent the wheel here) and then sent to the torrent client. There they are served via HTTP and the file pieces are downloaded as they are needed/streamed. A media player then sits in front of this web server and transcodes the media files in real time (rather than downloading different versions) for the web client. 

This means you can grab a single release at the highest quality/size you are willing, and transcode to meet your current internet speed wherever you are. No need for VPNs on your clients, all torrenting happens on your server/seedbox.

If you are still confused about why this project was started, check out the motivation section at the bottom... but also, I just had some free time and wanted to see how far I could take it :)

![Demo](demo-gif.gif)

## Dependencies
- [Jackett](https://github.com/Jackett/Jackett)
- [ffmpeg](https://ffmpeg.org/)

## Torrent client
Iceetime includes a fully featured torrent client so you can decide how you want the files to be downloaded and seeded (which helps solve issue 1 I mentioned above). We don't use existing clients because we specifically need the ability to serve files via HTTP and prioritize those streams over downloading the entire torrent.

Features:
- [x] Serves raw files via HTTP range requests which downloads pieces when they are needed
- [x] Add torrents via info hash
- [x] Add torrents via magnet url
- [x] Add torrents via file on disk
- [x] Find Movie files/torrents via Torznab queries
- [ ] Endpoint to check if a movie exists on disk already
- [ ] Find TV Shows/Episode files/torrents via Torznab queries
- [ ] Endpoint to check if a tv show or episode exists on disk already
- [ ] Download all pieces of a torrent when no one is streaming
- [ ] Web interface for managing torrents

## Media Player (realtime transcoder)
Iceetime also includes a layer on top of the raw files that aims to make your files as available as possible.

Features:
- [x] Transcode to different resolutions and bitrates
- [ ] Provide detailed metadata about files including all video/audio/subtitle tracks
- [ ] Transcode to different file formats
- [ ] Add subtitles during transcode
- [ ] Serve subtitle track so the client can decide if it wants to render them

## Web Client
The web client is fairly independent of the backend and aims to make it easy to select movies and then provide the backend the info it needs to go find a torrent for that movie.

Features:
- [x] Use TMDB api to get info on media
- [x] Request movies to be fetched
- [x] Stream movies
- [ ] Option to select transcode quality
- [ ] Page for movies with status about files on disk

# Docker Deployment

Building
```
pushd backend && make docker && popd
pushd frontend && make docker && popd
```

Running
```
docker run -it \
-v $(pwd)/config.yaml:/etc/config.yaml \
-v $(pwd)/dbs:/dbs \
-e CONFIG_FILE=/etc/config.yaml \
-e TORRENT_DB_FILE=/dbs/torrent.db \
-p 8080:8080
iceetime/backend

docker run -it \
-e REACT_APP_TMDB_API_KEY=/etc/config.yaml \
-e REACT_APP_TMDB_API_KEY=<your-api-key> \
-p 3000:3000 \
iceetime/frontend
```

Docker Compose

```
jackett:
    image: linuxserver/jackett
    container_name: jackett
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=America/Los_Angeles
    volumes:
      - jackett:/config
      - /downloads:/downloads
    ports:
      - 9117:9117
    restart: unless-stopped

iceetimeBackend:
    image: iceetime/backend
    container_name: iceetime-backend
    environment:
        - CONFIG_FILE=/etc/config.yaml
        - TORRENT_DB_FILE=/dbs/torrent.db
    volumes:
        - /mnt/iceetime/dbs:/dbs
        - /mnt/downloads:/downloads

iceetimeFrontend:
    image: iceetime/frontend
    container_name: iceetime-frontend
    environment:
        - REACT_APP_TMDB_API_KEY=<your-api-key>
```

# Motivation for this project (issues with Popcorntime)

Popcorntime is awesome for torrent usability, but has a few problems that make it a bit hard to use (for me). I think the easiest way to understand the motivation behind this project is to look at the problems I have with Popcorntime.

Remember that this isn't meant to bash their appliction! These are two completely different projects that tackle the problem of torrent streaming in totally different ways.

### 1. Hard to seed
PT has very little emphasis on seeding. You seed while you watch, but stops seeding when you close the app. This means there's no way you could use a private tracker, and in general you're just being a leech!

### 2. You need a VPN on all devices
Because each of the apps are actively torrenting, you end up needing to have a VPN on all of your devices you want to watch on. I don't usually like to have a VPN active on all my devices at all times and think it's a bit annoying to keep switching them on and off when I want to watch some shows.

### 3. Bad file availability
A smaller issue I noticed is that PTs solution to poor internet connectivity is to select a lower quality torrent rather than adjust the file you are downloading.
