# ESI WebSocket Proxy

Access to EVE Online's ESI without the hassle.
Using a WebSocket to retrieve data and events of any endpoint.

For example, subscribe to the `location` endpoint to only receive an event when your location changes.
No more polling every 5s yourself to find out your location hasn't changed!

## Work In Progress

This repository is a work-in-progress, and in fact, more a proof-of-concept.
Currently it only supports subscriptions to the location endpoint, and doesn't do anything further.

Additionally, when there are too many websockets subscribed to a location, it will hit the rate-limit of ESI.
There is currently no code to prevent that.

See [client/index.html](client/index.html) for more details why this repository exists, and where we are at.

## Usage

```bash
go run .
```

And navigate to http://localhost:8080.

Replace the `client_id` with one of your own if you want to use the SSO button.
