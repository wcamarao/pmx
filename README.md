# pmx

Golang data mapping library for postgres and pgx.

## Install

```
go get -u github.com/wcamarao/pmx
```

## Features

- Simple data mapping with struct tags
- Explicit by design, no magic or conventions
- Insert database records from an annotated struct
- Select database records into an annotated struct or slice
- Compatible with pgx Exec/Query interface i.e. `pgxpool.Pool`, `pgx.Conn`, `pgx.Tx`
- Support auto generated values, transient fields and pgx supported types e.g. `jsonb` maps and slices

## Data mapping

Given the following table:

```sql
create table events (
    id uuid primary key,
    payload jsonb,
    recorded_at timestamptz
);
```

Annotate a data model with struct tags:

```go
type Event struct {
    ID                string         `db:"id" table:"events"`
    Payload           map[string]any `db:"payload"`
    RecordedAt        time.Time      `db:"recorded_at"`
    ExportedTransient string         // ignored by pmx
    privateTransient  string         // ignored by pmx
}
```

- Annotate the first struct field with `table` (required for insert statements only)
- Annotate exported struct fields with `db` (required for insert and select statements)
- Exported fields without a `db` struct tag are transient and ignored by pmx
- Private fields are always transient and ignored by pmx

## Insert

Always provide a struct pointer.

Use the function `pmx.UniqueViolation(err)` to determine whether an insert statement failed due to unique violation.

```go
func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgresql://...")
    if err != nil {
        log.Panicf("connection error: %+v", err)
    }
    defer conn.Close(ctx)

    event := Event{
        ID:         "a1eff19b-4624-46c6-9e09-5910e7b2938d",
        Payload:    map[string]any{"key", "value"},
        RecordedAt: time.Now(),
    }

    // Generated query: insert into events (id, payload, recorded_at) values ($1, $2, $3)
    tag, err := pmx.Insert(ctx, conn, &event)
    if pmx.UniqueViolation(err) {
        log.Panicf("unique violation error: %+v", err)
    }
    if err != nil {
        log.Panicf("error: %+v", err)
    }

    log.Printf("tag: %+v", tag)
    log.Printf("event: %+v", event)
}
```

## Select one record

Always provide a struct pointer.

Selecting one record will error with `pmx.ErrNoRows` if an empty dataset is returned from the query.

```go
func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgresql://...")
    if err != nil {
        log.Panicf("connection error: %+v", err)
    }
    defer conn.Close(ctx)

    var event Event
    err = pmx.Select(ctx, conn, &event, "select * from events where id = $1", "a1eff19b-4624-46c6-9e09-5910e7b2938d")
    if errors.Is(err, pmx.ErrNoRows) {
        log.Panicf("event not found error: %+v", err)
    }
    if err != nil {
        log.Panicf("error: %+v", err)
    }

    log.Printf("event: %+v", event)
}
```

## Select multiple records

Always provide a slice pointer. The underlying slice type must be a struct pointer.

Selecting multiple records won't error if an empty dataset is returned from the query.

```go
func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgresql://...")
    if err != nil {
        log.Panicf("connection error: %+v", err)
    }
    defer conn.Close(ctx)

    var events []*Event
    err = pmx.Select(ctx, conn, &events, "select * from events limit 10")
    if err != nil {
        log.Panicf("error: %+v", err)
    }
    if len(events) == 0 {
      log.Panicf("no events found: %+v", events)
    }

    log.Printf("events: %+v", events)
}
```

## Auto generated values

Given the following table with auto generated values:

```sql
create table events (
		id bigserial primary key,
		recorded_at timestamptz default now()
);
```

Annotate struct fields with the `generated:"auto"` struct tag:

```go
type Event struct {
    ID         int64     `db:"id"          generated:"auto" table:"events"`
    RecordedAt time.Time `db:"recorded_at" generated:"auto"`
}
```

- Both columns will use `default` (aka auto generated) values in `insert` statements.
- Both struct fields will be populated with values from `insert returning` statements.

## ErrInvalidRef

The error `pmx.ErrInvalidRef` ("invalid ref") means you provided an invalid pointer reference.

## ErrNoRows

The error `pmx.ErrNoRows` means an empty dataset was returned when trying to select one record.

This error is a reference to the underlying `pgx.ErrNoRows` for convenience in data access layers.
