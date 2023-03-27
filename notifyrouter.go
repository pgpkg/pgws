package pgws

import "sync"

// NotifyRouter allows websocket listeners to register themselves
// to receive notifications. Notifications are delivered to individual
// websockets based on team ID. There might be multiple websockets
// registered for a single team ID, and that's OK.

type NotifyRouter struct {
	registrations map[string][]MessagePoster
	mutex         sync.Mutex
}

// MessagePoster is an interface which is can be implemented by any
// object that wants to listen in on messages posted from PG. The expectation
// is that posting the message itself will be done robustly by the websocket,
// for example by maintainin and internal queue of messages and dropping
// connections which can't keep up.
type MessagePoster interface {
	Post(message []byte)
}

func NewNotifyRouter() *NotifyRouter {
	return &NotifyRouter{
		registrations: make(map[string][]MessagePoster),
	}
}

// Post posts a message to all registered channels for a team. It's OK if
// there are no channels registered for a given team, it just means
// nobody is listening at the moment.
func (r *NotifyRouter) Post(team string, message []byte) {
	r.mutex.Lock()
	posters := r.registrations[team]
	r.mutex.Unlock()

	if posters == nil {
		return
	}

	for _, poster := range posters {
		poster.Post(message)
	}
}

// Register registers a new MessagePoster for the given team.
func (r *NotifyRouter) Register(team string, p MessagePoster) {
	r.mutex.Lock()
	r.registrations[team] = append(r.registrations[team], p)
	r.mutex.Unlock()
}

// Unregister removes a MessagePoster from the given team.
func (r *NotifyRouter) Unregister(team string, p MessagePoster) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	current := r.registrations[team]
	var next []MessagePoster

	for _, poster := range current {
		if poster != p {
			next = append(next, poster)
		}
	}

	r.registrations[team] = next
}
