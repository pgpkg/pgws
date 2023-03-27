# PGWS

`pgws` is a small Go library which lets you send messages directly to a WebSocket
connection through the use of Postgresql's NOTIFY statement.

NOTIFY (and its functional equivalent, `pg_notify`) is a great way to communicate
between the database and a web client because it's transactional: the NOTIFY
only completes if the transaction is successfully committed, which means that
your NOTIFY commands can be interspersed with other SQL commands, or can be
included in triggers.

`pgws` is multi-tenanted by design. Only messages intended for a particular websocket
client will be delivered.

## Using PGWS

The following WS server can be found in [cmd/pgws/main.go](cms/pgws.main.go):

    package main
    
    import (
        "github.com/bookworkhq/pgws"
        "log"
        "net/http"
        "time"
    )
    
    func main() {
        pgws := pgws.StartPGWebSocket("", 10*time.Second, time.Minute, "pgwebsocket")
        http.Handle("/ws", pgws)
        log.Fatal(http.ListenAndServe(":8080", nil))
    }

The client-side looks like this:

    <!doctype html>
    
    <html>
    <head>
        <script>
            const socket = new WebSocket('ws://localhost:8080/ws');
    
            socket.addEventListener('open', event => {
                console.log('WebSocket connection opened');
                // socket.send('Hello, server!');
            });
    
            socket.addEventListener('message', event => {
                console.log('Received message:', JSON.parse(event.data));
            });
    
            socket.addEventListener('close', event => {
                console.log('WebSocket connection closed');
            });
        </script>
    </head>
    </html>

To send a message to the client from psql:

    sql> select pg_notify('pgwebsocket', 'test-team,{"message": "test at ' || current_timestamp || '"}');
    pg_notify
    -----------
    
    (1 row)

The client will print the message on the console.

Note that the message is assumed to be JSON, and this will probably be enforced in future
versions.