### Anacrolix Torrent Abstraction

This package acts as a very simple abstraction to anacrolix's torrent package. The main goal is to just let us use the `Torrent` and `Client` objects as **interfaces** in our `/app` business logic.

This is very powerful and allows us to mock those objects out and easily unit test the actual logic instead of having to write tests that actually mess with the torrent client.