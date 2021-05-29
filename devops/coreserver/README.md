# Core Server

The core server handles a few things:

1. Handling requests for integration tests and forwarding them to the integration test server.
2. Sending heartbeat mesages to servers and restarting them if necessary.
3. Sending requests to relevant servers when the main repo is updated.
