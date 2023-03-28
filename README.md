# PGWS

`pgws` is a small Go library which lets you send messages directly to a browser via
a persistent WebSocket connection, using Postgresql's NOTIFY statement.

NOTIFY (and its functional equivalent, `pg_notify`) is a great way to communicate
between a database and a web client because it's transactional: any NOTIFY
commands issued during a transaction only get sent if the transaction is
successfully committed - but you don't need a database table to make it
work. This makes them perfect for telling a web client that a database row
has been updated, or even to push the row directly to the client.

`pgws` allows you listen for NOTIFY messages from any number of channels.

## Status

`pgws` is early alpha. See the [TODO](TODO.md) file for a list of issues.

## Features

* Easy to use. Go Setup is two lines of code; pushing a message from pgsql is just one.
* Works with any Postgres server. No extensions or logical replication slots required.
* Works with Go's `http` package; implements `http.Handler`
* Efficient. Uses a single database connection for all websockets.
* [Audience filters](#audiences) - provides security for multi-tenant applications.

PGWS uses `fasthttp/websocket` for the websocket component.

## Using PGWS

### Server side

A minimal WebSocket server can be found in [cmd/pgws/main.go](cms/pgws.main.go):

    package main
    
    import (
      "github.com/pgpkg/pgws"
      "log"
      "net/http"
      "time"
    )
    
    func main() {
        // Create a listener on the PG database.
        l := pgws.StartPGListener("", 10*time.Second, time.Minute)
      
        // Create a websocket endpoint associated with the listener.
        // This one is listening on the NOTIFY channel "pgws". You can specify
        // any number of channels.
        endpoint := pgws.NewPGWS(l, "pgws")
    
        // Add it to the default router...
        http.Handle("/ws", endpoint)
    
        // ...and start the HTTP server.
        log.Fatal(http.ListenAndServe(":8080", nil)) 
    }

### Client Side

A minimal client that logs to the console looks like this:

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

> Note: To prevent a client from joining an [audience](#audiences) that it isn't entitled to,
> the audience for each client needs to be determined on the **server side**,
> perhaps through the use of cookies, JWTs etc. See the link for more information.

## Sending Messages

You send messages using the `NOTIFY` command or the `pg_notify()` function.
Note that `pg_notify` is more flexible.

The message itself should be a text string in the following format:

    audience,{ ... }

where `audience` is the [audience](#audiences) that you're sending the message to, and the remainder
of the message is a JSON object. This simple scheme means that we don't have to parse the JSON
to post the message.

To send a message to the client from psql:

    sql> select pg_notify('pgws', 'default,{"message": "hello, world"}');
    pg_notify
    -----------
    
    (1 row)

The simple client should then print the message on the console.

To use the built-in JSON functions in Postgresql, you could do something
like this:

    select pg_notify('pgws', 'default,' || jsonb_build_object('message', 'hello, world')::text);

`pgws` requires that the message payload is a well-formed JSON object.

## Message format

You can post any JSON using the NOTIFY command or `pg_notify` function.

However, the message posted to a client is not identical to the one received from NOTIFY.
`pgws` parses the JSON (to ensure it's well-formed), and puts it inside a wrapper
which provides metadata about the message to the client. Note that this process might also
change the order of fields in the JSON object delivered to the client.

The result of the `pg_notify` example above would look something like this:

    {
        "payload": {
            "message": "hello, world"
        },
        "channel": "pgws",
        "id": "b8cfabc9-6472-46a4-babd-b575ce6433cf"
    }

The message sent from NOTIFY is included in `payload`; `pgws` adds metadata in
the form of the `channel` name and a unique `id`. The `id` is associated with the
message itself; the same ID is provided to all clients, which allows you to correlate messages
during testing.

Note that the [audience](#audiences) is not provided in the metadata 

## Audiences

In `pgws`, an "audience" identifies the group of users for which a message is intended.
Audiences could be used in a number of ways, but the intent was to enable efficient
delivery of messages in multi-tenant applications. Members of different audiences
should not be able to see one another's messages.

If you don't do anything, the only configured audience will be "default", and all
clients will receive all messages sent to the "default" audience.

Becasue we can't trust clients, the audience for an incoming websocket connection
must be determined at the time of the initiating request. That is, when the browser
performs a GET on the `pgws` server URL, your server needs to determine which audience
that connection will belong to.

To do this, set the GetAudience field of the PGWS struct to a function that returns it:

	l := pgws.StartPGListener("", 10*time.Second, time.Minute)
	endpoint := pgws.NewPGWS(l, "pgws")
    endpoint.GetAudience = func(r *http.Request) { return []string{"noosa", "coolum"} }

In this example, messages sent to the "default" audience would no longer be delivered,
but messages sent to the "noosa" and "coolum" audiences would be sent to everyone.

Of course, instead of returning a literal, the intent is that you would use the 
GetAudience function to introspect the `http.Request` object to determine
the tenant ID or other attribute of the user (say, via a JWT, cookies, the URL, or a request
header injected by a reverse proxy), to determine which audience they belong to.

Audiences are specifically designed to allow messages to be sent to selected
websockets quickly and efficiently.

## Security

`pgws` doesn't provide any security at all. The intent is that you would wrap the
PGWS server with an authorization function, such as a JWT validator, if that's what
you need.