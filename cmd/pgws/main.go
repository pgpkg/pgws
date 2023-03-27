package main

import (
	"github.com/pgws/pgws"
	"log"
	"net/http"
	"time"
)

func main() {
	pgws := pgws.StartPGWebSocket("", 10*time.Second, time.Minute, "pgwebsocket")
	http.Handle("/ws", pgws)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
