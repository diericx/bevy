### `/repos`

This directory contains small intermediary structs for reading and storing data called repos. They are sepperated by storage location (jackett and storm). They should contain no business logic and only act as a tool to interact with persistant.

Jackett is a weird one here. We are pretending Jackett is some read only database we are fetching from via queries.