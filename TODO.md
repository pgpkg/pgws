# pgnotify todo

- [ ] The MessagePoster.Post() WebSocket implementation can probably wait forever.
      needs better monitoring.
- [ ] APIs to support JWT integration
  - [ ] time out connection when JWT expires (maximum lifetime for a single connection)
  - [ ] filter by team ID (correlate with JWT `team` claims?) - list of teams against connection
