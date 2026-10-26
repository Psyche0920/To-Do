# Backend Learning Notes

## Requests and responses

- `w http.ResponseWriter`: `w` is the parameter name; `http.ResponseWriter` is its interface type. It provides methods for writing a response; it is not JSON.
- `r *http.Request`: `r` points to the incoming request. `r.Body` contains the client's request payload.
- `Header().Set` sets response metadata. `WriteHeader` sends the status code and headers. JSON encoding writes the response body.
- A handler writes through `w`; it does not return a response object as a FastAPI handler commonly does.

## Application and database

- `application` is our dependency-holding struct. `app` points to it; `app.db` points to the shared connection pool.
- A method receiver `(app *application)` gives the method access to that instance, similar to Python's `self`.
- `sql.DB` is the pool type; `*sql.DB` is a pointer to a pool. PostgreSQL is the separate database server.
- DSN is connection information. The pgx driver handles communication. `sql.Open` prepares the pool; `PingContext` checks connectivity.
- SQL is the instruction language; PostgreSQL is the database software that executes it.

## Creating a task

- Input accepts `title` only. PostgreSQL generates `id` and defaults `completed` to `false`.
- `NewDecoder(r.Body)` prepares a JSON reader. `Decode(&input)` reads JSON into Go data; it does not convert data into JSON.
- `DisallowUnknownFields` rejects extra fields inside the object. A second decode expecting `io.EOF` rejects extra content after the object.
- Two objects in one request body are different from two separate HTTP requests. Clicking twice usually sends two requests; this check does not prevent duplicate submissions.
- `1<<20` is 1 MiB (1,048,576 bytes), the maximum body size, not a required size.
- `TrimSpace` removes surrounding whitespace. `len([]rune(title))` counts Unicode code points, not bytes; complex visible characters can contain multiple code points.
- `QueryRowContext` executes SQL that returns one row. Here, `INSERT ... RETURNING` inserts a task and returns its fields.
- `Scan` copies returned columns into the supplied variable addresses, in column order. Its error can come from execution or reading results; an error alone does not prove an insert never happened.

## Timeouts and cleanup

- `context` is a package; `Background()` creates a base context. `r.Context()` carries the HTTP request's cancellation.
- `WithTimeout` starts a deadline when called. Only operations using the resulting context follow that deadline; later code is not automatically limited.
- The create handler passes `ctx` to `QueryRowContext`. Reading its returned row with `Scan` uses the same operation and remaining deadline; it does not get another five seconds.
- This deadline covers database work, including waiting for a connection, executing SQL and reading results. It does not cover earlier JSON decoding or later response writing.
- `cancel()` releases context resources. `rows.Close()` releases result resources. `db.Close()` closes the connection pool.

## Errors

- `nil` means no error when used as an error result. Retrieve the error first, then check `err != nil`.
- `fmt.Errorf` creates an error; it does not print or save it. `%w` preserves a wrapped error; `%v` formats a value for display.
- Successful database setup returns `db, nil`; failure returns `nil, err`. Both return positions are always present.
- `log.Printf` writes to the API process's logs. HTTP errors must be written separately for the client.
