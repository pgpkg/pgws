# pgws todo

- [ ] The MessagePoster.Post() WebSocket implementation will probably wait forever.
      needs better monitoring.
- [ ] APIs to support JWT integration
  - [ ] time out connection when JWT expires (maximum lifetime for a single connection)
  - [ ] filter by audience ID for multi-tenancy
