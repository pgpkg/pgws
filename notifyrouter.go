package pgws

// NotifyRouter allows websocket listeners to register themselves
// to receive notifications. Notifications are delivered to individual
// websockets based on the audience. There might be multiple websockets
// registered for a single audience, and that's OK.

type NotifyRouter struct {
	registrations map[string][]MessagePoster
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

// Post posts a message to all registered channels for a audience. It's OK if
// there are no channels registered for a given audience, it just means
// nobody is listening at the moment.
func (r *NotifyRouter) Post(audience string, message []byte) {
	posters := r.registrations[audience]

	if posters == nil {
		return
	}

	for _, poster := range posters {
		poster.Post(message)
	}
}

// Register registers a new MessagePoster for the given audience.
func (r *NotifyRouter) Register(audiences []string, p MessagePoster) {
	for _, audience := range audiences {
		r.registrations[audience] = append(r.registrations[audience], p)
	}
}

// Unregister removes a MessagePoster from the given audience.
func (r *NotifyRouter) Unregister(audiences []string, p MessagePoster) {
	// This could certainly be a bit more efficient.
	for _, audience := range audiences {
		current := r.registrations[audience]
		var next []MessagePoster

		for _, poster := range current {
			if poster != p {
				next = append(next, poster)
			}
		}
		r.registrations[audience] = next
	}
}
