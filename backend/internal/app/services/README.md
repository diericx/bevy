### `/services`

Services use repos and our domain in the `app` package to apply our business logic.
These will be much beefier than the repos and will contain actual logic. We want to abstract as much of the structs we inject here (repos and torrent client) so we can perform real unit tests on this logic. We want this dir more than any to be completely sound.