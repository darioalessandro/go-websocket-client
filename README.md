# go-websocket-client

## 1fireandforgetwebsocket

Establishes a websocket connection to the url defined, sends ping messages every second.

If the connection breaks, it does not attempt to reconect.

## 2resilientwebsocketclient

Implements a supervisor routine that is used to detect when the websocket connects/disconnects and has a
infinite retry logic to handle disconnection events.

