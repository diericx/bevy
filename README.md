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

Currently we are using the default Go HTTP Range function which just serves a file using implementations of functions the html5 player expects. It makes some assumptions about the file which don't match our use case.

If we reimplement this protocol, we might be able to transcode live without having to do much fluff like dealing with metadata files and what not... we just need to fully understand HTTP Range requests.