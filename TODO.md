# pgws todo

- [ ] websocket.Upgrader removes origin checks. That's bad!
- [ ] removing listeners needs to be more performant, it's currently O(n).
- [ ] The MessagePoster.Post() WebSocket implementation will probably wait forever.
      needs better monitoring, queues, disconnects.
- [ ] APIs to support JWT integration
  - [ ] timeout websocket when JWT expires (ie, set maximum lifetime for a single connection)
- 
- [X] filter by audience ID for multi-tenancy
- [X] need to include the channel in the message delivered to the client. (json)
