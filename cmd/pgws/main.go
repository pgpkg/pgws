package main

import (
	"github.com/pgpkg/pgws"
	"log"
	"net/http"
	"time"
)

func main() {
	// Create a single listener on the PG database.
	l := pgws.StartPGListener("", 10*time.Second, time.Minute)

	// Create a websocket endpoint associated with the listener.
	// This one is listening on the NOTIFY channel "pgws". You can specify
	// any number of channels.
	endpoint := pgws.NewPGWS(l, "pgws")

	// Add it to the default router...
	http.Handle("/ws", endpoint)

	// ...and start the HTTP server.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
