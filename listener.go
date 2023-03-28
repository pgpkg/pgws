package pgws

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"strings"
	"sync"
	"time"
)

type PGListener struct {
	pqListener *pq.Listener
	routers    map[string]*NotifyRouter
	mutex      sync.Mutex
}

// Message is the decomposed message as received from NOTIFY
// that we will send to the client. It needs to include the channel,
// and we provide an ID for all messages to help with correlation.
type Message struct {
	audience string         // this is not sent to the client.
	Payload  map[string]any `json:"payload"` // the JSON message
	Channel  string         `json:"channel"`
	ID       string         `json:"id"` // random UUID used to correlate messages across clients
}

func decode(n *pq.Notification) (*Message, error) {
	msg := n.Extra
	sep := strings.Index(msg, ",{")

	// Separator must exist and can only be as long as a UUID.
	if sep == -1 || sep > 36 {
		return nil, fmt.Errorf("invalid or missing separator in message", msg)
	}

	message := Message{
		audience: msg[:sep],
		Channel:  n.Channel,
		ID:       uuid.New().String(),
		//Payload:  make(map[string]any),
	}

	if err := json.Unmarshal([]byte(msg[sep+1:]), &message.Payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %w", err)
	}

	return &message, nil
}

// listen listens for NOTIFY messages of the form "audience,{...}"
// and posts them to listening websockets. This function doesn't return,
// and will continue to attempt to connect to the database on failure.
func (l *PGListener) listen() {
	for n := range l.pqListener.Notify {
		if n == nil {
			// See https://pkg.go.dev/github.com/lib/pq@v1.10.7#hdr-Notifications
			// Nil is sent if there is a reconnect. It's not clear how to deal with this
			// in pgws yet, since all channels (and all audiences are affected).
			continue
		}

		message, err := decode(n)
		if err != nil {
			log.Printf("unable to decode message: %v", err)
			continue
		}

		mb, err := json.Marshal(message)
		if err != nil {
			log.Printf("unable to encode message: %v", err)
			continue
		}

		// Find the router for this channel, if one exists
		router := l.routers[n.Channel]
		if router != nil {
			router.Post(message.audience, mb)
		}
	}
}

// AddPoster adds a poster (websocket) to the listener. If the channels are not already
// being listened to, they are added.
func (l *PGListener) AddPoster(pgChannels []string, audiences []string, p *WSPoster) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Add the Poster to the router for each channel.
	for _, pgChannel := range pgChannels {
		n := l.routers[pgChannel]
		if n == nil {
			n = NewNotifyRouter()
			l.routers[pgChannel] = n
			if err := l.pqListener.Listen(pgChannel); err != nil {
				panic(err)
			}
		}
		n.Register(audiences, p)
	}
}

// RemovePoster removes a poster (websocket) from the listener.
func (l *PGListener) RemovePoster(pgChannels []string, audiences []string, p *WSPoster) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Add the Poster to the router for each channel.
	for _, pgChannel := range pgChannels {
		n := l.routers[pgChannel]
		if n != nil {
			n.Unregister(audiences, p)
		}
	}
}

// StartPGListener starts (and returns) a listener on the PG database.
func StartPGListener(dsn string, minReconn time.Duration, maxReconn time.Duration) *PGListener {
	l := &PGListener{
		pqListener: pq.NewListener(dsn, minReconn, maxReconn, pqlCallback),
		routers:    make(map[string]*NotifyRouter),
	}
	go l.listen()
	return l
}
