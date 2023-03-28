package pgws

import (
	"github.com/fasthttp/websocket"
	"net/http"
)

type WSPoster struct {
	PGWS      *PGWS
	Conn      *websocket.Conn
	closeChan chan bool
}

func (poster *WSPoster) Post(message []byte) {
	err := poster.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		poster.Conn.Close()
	}
}

// Read (and ignore) all messages from the connection. This is
// necessary in order to receive close notifications.
func (poster *WSPoster) readAll() {
	if _, _, err := poster.Conn.NextReader(); err != nil {
		poster.Conn.Close()
	}
}

func (poster *WSPoster) Close(code int, text string) error {
	poster.closeChan <- true
	return nil
}

func (poster *WSPoster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer poster.Conn.Close()

	pgws := poster.PGWS
	audiences := pgws.getAudience(r) // Only use the private version of getAudience here.
	poster.Conn.SetCloseHandler(poster.Close)
	pgws.Listener.AddPoster(pgws.PGChannels, audiences, poster)

	// We don't care about incoming messages from the websocket, but we must
	// perform reads in order to receive close messages. So we do this in a loop.
	// See https://pkg.go.dev/github.com/fasthttp/websocket#hdr-Control_Messages
	go poster.readAll()

	select {
	// TODO: we also need to terminate when the JWT expires, ie. by setting a timer.
	case <-poster.closeChan:
		break
	}

	pgws.Listener.RemovePoster(audiences, pgws.PGChannels, poster)
}
