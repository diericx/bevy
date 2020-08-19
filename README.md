# iceetime

Iceetime is a self hosted, isolated alternative to Popcorntime that aims to follow torrent ethos and be private tracker friendly. 

You host an instance of iceetime on your server, and can then stream content from it to any device. 
The server is the only device that needs a VPN as it's the only one torrenting. 

## Current state of the project

This project can currently download a torrent sequentially and serve that stream to a browser via Gos built in servecontent function.

It is missing:
- Torrent indexing (searching for torrents by quality)
- Transcoding

I think indexing can either be solved by simply using Sonarr/Radarr although I really don't like using those programs... shitty UI/UX that should be configs so I can actually see whats happening. Not very developer friendly.

## Option 1. Transcoding solved with MPEG DASH

It's possible to solve the transcoding problem by implementing MPEG DASH. Essentially the process would look like this:

Client requests section 1 -> server starts downloading that section -> server starts encoding that section -> server serves transcoded section to client

But the client only requests sections it knows about from a metadata file (`.mpd` file) that is generated when you convert a media file to DASH format. We would need to either generate these manually to spoof it or find some other way to solve this problem.

## Option 2. Transcoding with custom HTTP Range implementation

Currently we are using the [default Go HTTP Range function `ServeContent`](https://golang.org/pkg/net/http/#ServeContent) which just serves a file using implementations of functions the html5 player expects. It makes some assumptions about the file which don't match our use case.

If we reimplement this protocol, we might be able to transcode live without having to do much fluff like dealing with metadata files and what not... we just need to fully understand HTTP Range requests.

I think it might be this easy...

The first thing this implementation does is [checks how large the file in the readseeker is by seeking to the end](https://github.com/golang/go/blob/ba9e10889976025ee1d027db6b1cad383ec56de8/src/net/http/fs.go#L157) and then making assumptions on that... which works when using raw files as we can see in the working version in this repo. What we need to do is add a step which checks how large the file downloading is, and then calculates the size the transcoded file will be. Then all operations should request correct byte sections and we can intercept requests by implementing our own io.readseeker and transcode as they come from the download readseeker then serve... hacky!

# Notes

Exaple query using Jackett
`192.168.1.71:9117/api/v2.0/indexers/torrentleech/results/torznab/api?apikey=0x7ym4k6c4nghc6nh6qi3s2pdyicxj19&t=movie&imdbid=tt0317705&cat=2040`

Note: On stupud ass mac run this
```
export CPATH="/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/usr/include/"
export CGO_ENABLED=1; export CC=gcc;
```

Disable annoying cgo warnings on mac
```
export CGO_CPPFLAGS="-Wno-nullability-completeness"
```