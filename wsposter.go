package pgws

import (
	"github.com/fasthttp/websocket"
	"net/http"
)

type WSPoster struct {
	Conn      *websocket.Conn
	closeChan chan bool
}

func (ws *WSPoster) Post(message []byte) {
	err := ws.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		ws.Conn.Close()
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Read (and ignore) all messages from the connection. This is
// necessary in order to receive close notifications.
func (ws *WSPoster) readAll() {
	if _, _, err := ws.Conn.NextReader(); err != nil {
		ws.Conn.Close()
	}
}

func (ws *WSPoster) Close(code int, text string) error {
	ws.closeChan <- true
	return nil
}

func HandleWebsocketPoster(audience []string, conn *websocket.Conn, n *NotifyRouter) {

	poster := WSPoster{
		Conn:      conn,
		closeChan: make(chan bool),
	}

	conn.SetCloseHandler(poster.Close)

	n.Register(audience, &poster)

	// We don't care about incoming messages from the websocket, but we must
	// perform reads in order to receive close messages. So we do this in a loop.
	// See https://pkg.go.dev/github.com/fasthttp/websocket#hdr-Control_Messages
	go poster.readAll()

	select {
	// TODO: we also need to terminate when the JWT expires, ie. by setting a timer.
	case <-poster.closeChan:
		break
	}

	conn.Close()
}
