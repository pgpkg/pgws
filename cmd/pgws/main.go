package main

import (
	"github.com/bookworkhq/pgws"
	"log"
	"net/http"
	"time"
)

func main() {
	pgws := pgws.StartPGWebSocket("", 10*time.Second, time.Minute, "pgwebsocket")
	http.Handle("/ws", pgws)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
