# PGWS

`pgws` is a small Go library which lets you send messages directly to a browser via
a persistent WebSocket connection, using Postgresql's NOTIFY statement.

NOTIFY (and its functional equivalent, `pg_notify`) is a great way to communicate
between a database and a web client because it's transactional: any NOTIFY
commands issued during a transaction only get sent if the transaction is
successfully committed, but you don't need a database table to make it
happen. This makes them perfect for telling a web client that a database row
has been updated, or even to push the row directly to the client.

## Status

`pgws` is early alpha. See the [TODO](TODO.md) file for a list of issues.

## Features

* Easy to use. Go Setup is a single line of code; so is pushing a message from pgsql.
* Works with Go's `http` package; implements `http.Handler`
* Audience filters - ideal for multi-tenant applications (see below).
* Single-connection. `pgws` makes only a single database connection regardless of
  the number of websockets it's serving.

PGWS uses `fasthttp/websocket` for the websocket component.

## Using PGWS

### Server side

A minimal WebSocket server can be found in [cmd/pgws/main.go](cms/pgws.main.go):

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

### Client Side

A minimal client that logs ot the console looks like this:

    <!doctype html>
    
    <html>
    <head>
        <script>
            const socket = new WebSocket('ws://localhost:8080/ws');
            socket.addEventListener('message', event => {
                console.log(event.data);
            });
        </script>
    </head>
    </html>

A (slightly) more interesting example can be found in [client.html](client.html).

## Sending Messages

You send messages using the `NOTIFY` command or the `pg_notify()` function.
Note that `pg_notify` is more flexible.

The message itself should be a text string in the following format:

    audience,{ ... }

where `audience` is the [audience](#audiences) that you're sending the message to, and the remainder
of the message is a JSON object. This simple scheme means that we don't have to parse the JSON
to post the message.

To send a message to the client from psql:

    sql> select pg_notify('pgwebsocket', 'default,{"message": "test at ' || current_timestamp || '"}');
    pg_notify
    -----------
    
    (1 row)

The client will print the message on the console.

Note that the message is assumed to be JSON, but we currently just copy the object
verbatim.

## Audiences

In `pgws`, an "audience" identifies the users for whom a message is intended.
Audiences could be used in a number of ways, but the intent was to enable efficient
delivery of messages in multi-tenant applications. If you don't do anything,
the only configured audience will be "default"; everyone will receive all messages
to the "default" audience.

The audience for a websocket is determined at the time of the initiating HTTP request.
To change the audience for a socket, set the GetAudience field of the PGWS struct to
a function that returns it:

	pgws := pgws.StartPGWebSocket("", 10*time.Second, time.Minute, "pgwebsocket")
    pgws.GetAudience = func(r *http.Request) { return []string{"noosa", "coolum"} }

In this example, messages sent to the "default" audience would no longer be delivered,
but messages sent to the "noosa" and "coolum" audiences would be sent to everyone.

Of course, the intent is that you would use the `http.Request` object to determine
the tenant ID or other attribute of the user (say, via a JWT, the URL, or a request
header injected by your proxy), to determine which audience they belong to.

## Security

`pgws` doesn't provide any security at all. The intent is that you would wrap the
PGWS server with an authorization function, such as a JWT validator, if that's what
you need.