# pgws todo

- [ ] websocket.Upgrader removes origin checks. That's bad!
- [ ] The MessagePoster.Post() WebSocket implementation will probably wait forever.
      needs better monitoring, queues, disconnects.
- [ ] APIs to support JWT integration
  - [ ] timeout websocket when JWT expires (ie, set maximum lifetime for a single connection)
  - [X] filter by audience ID for multi-tenancy
