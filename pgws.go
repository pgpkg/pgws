package pgws

import (
	"fmt"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"time"
)

type PGWS struct {
	notifyRouter *NotifyRouter
	pqListener   *pq.Listener
	PGChannel    string // The PG channel to LISTEN on
}

func pqlCallback(ev pq.ListenerEventType, err error) {
	if err != nil {
		fmt.Printf("pgwebsocket: WARNING: %s\n", err.Error())
	}
}

// listen listens for NOTIFY messages of the form "team,{...}"
// and posts them to listening websockets.
func (pgws *PGWS) listen() {
	err := pgws.pqListener.Listen(pgws.PGChannel)
	if err != nil {
		// There's no coming back from this since we're in a goroutine.
		// Better to explode than fail silently. We only have one job!
		panic(err)
	}

	for n := range pgws.pqListener.Notify {
		msg := n.Extra
		sep := strings.Index(msg, ",{")

		// Separator must exist and can only be as long as a UUID.
		if sep == -1 || sep > 36 {
			log.Println("invalid or missing separator in message", msg)
			continue
		}

		team := msg[:sep]
		payload := msg[sep+1:]

		pgws.notifyRouter.Post(team, []byte(payload))
	}
}

func (pgws *PGWS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	HandleWebsocketPoster("test-team", conn, pgws.notifyRouter)
}

func StartPGWebSocket(dsn string, minReconn time.Duration, maxReconn time.Duration, pgChannel string) *PGWS {
	pgws := &PGWS{
		PGChannel:    pgChannel,
		notifyRouter: NewNotifyRouter(),
		pqListener:   pq.NewListener(dsn, minReconn, maxReconn, pqlCallback),
	}
	go pgws.listen()
	return pgws
}
