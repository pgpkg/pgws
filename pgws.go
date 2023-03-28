package pgws

import (
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/lib/pq"
	"log"
	"net/http"
)

// PGWS represents a single websocket endpoint for a given set of
// PG channels. You only need to create a single PG connection for a given DSN, regardless
// of the number of websocket endpoints you create.
//
// PGWS implements http.Handler

type PGWS struct {
	PGChannels  []string                       // The PG channels to LISTEN on
	Listener    *PGListener                    // The listener associated with this WS endpoint
	GetAudience func(r *http.Request) []string // returns the audiences for this connection (defaults to 'default')
}

func pqlCallback(ev pq.ListenerEventType, err error) {
	if err != nil {
		fmt.Printf("pgwebsocket: WARNING: %s\n", err.Error())
	}
}

// getAudience identifies the audience(s) for a WebSocket based on the original HTTP request.
// Different sockets will receive a different subset of messages depending on the audience
// they belong to. How the audience is determined is decided when the socket connects.
func (pgws *PGWS) getAudience(r *http.Request) []string {
	if pgws.GetAudience != nil {
		return pgws.GetAudience(r)
	}
	return []string{"default"}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// ServeHTTP accepts incoming HTTP connections, upgrades them to WebSockets,
// and serves the websocket. It only returns when the WebSocket closes.
// Note that there will be many calls to ServeHTTP for the same instance of PGWS.
// A new instance of WSPoster is created for each HTTP connection we upgrade.
func (pgws *PGWS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	poster := WSPoster{
		PGWS:      pgws,
		Conn:      conn,
		closeChan: make(chan bool),
	}
	poster.ServeHTTP(w, r)
}

func NewPGWS(listener *PGListener, pgChannels ...string) *PGWS {
	return &PGWS{
		Listener:   listener,
		PGChannels: pgChannels,
	}
}
