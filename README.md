# pmx

Golang data mapping library for postgres and pgx.

## Install

```
go get -u github.com/wcamarao/pmx
```

## Features

- Simple data mapping with struct tags
- Explicit by design, no magic or conventions
- Select database records into an annotated struct or slice
- Insert and update database records from an annotated struct
- Compatible with pgx Exec/Query interface i.e. `pgxpool.Pool`, `pgx.Conn`, `pgx.Tx`
- Support transient fields and auto generated values

## Data Mapping

Given the following table:

```sql
create table events (
    id uuid primary key,
    recorded_at timestamptz
);
```

Annotate a data model with struct tags:

```go
type Event struct {
    ID         string    `db:"id" table:"events"`
    RecordedAt time.Time `db:"recorded_at"`
    Transient  string    // ignored by pmx
    transient  string    // ignored by pmx
}
```

- The _first_ struct field must be annotated with `table` for insert and update operations
- Struct fields annotated with `db` must be exported
- Transient fields can be optionally exported

## Insert

Always provide a struct pointer.

```go
type Event struct {
    ID         string    `db:"id" table:"events"`
    RecordedAt time.Time `db:"recorded_at"`
}

func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgresql://...")
    if err != nil {
        panic(err)
    }
    defer conn.Close(ctx)

    event := Event{
        ID:         "a1eff19b-4624-46c6-9e09-5910e7b2938d",
        RecordedAt: time.Now(),
    }

    // Generated query: insert into events (id, recorded_at) values ($1, $2)
    _, err = pmx.Insert(ctx, conn, &event)
    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", event)
}
```

## Select into struct

Always provide a _struct_ pointer.

Selecting into a _struct_ will error with `pgx.ErrNoRows` if an empty dataset is returned from the query.

```go
type Event struct {
    ID         string    `db:"id" table:"events"`
    RecordedAt time.Time `db:"recorded_at"`
}

func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgresql://...")
    if err != nil {
        panic(err)
    }
    defer conn.Close(ctx)

    var event Event
    err = pmx.Select(ctx, conn, &event, "select * from events where id = $1", "a1eff19b-4624-46c6-9e09-5910e7b2938d")
    if errors.Is(err, pgx.ErrNoRows) {
        panic("event not found")
    }
    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", event)
}
```

## Select into slice

Always provide a _slice_ pointer. The underlying slice type must be a _struct_ pointer.

Selecting into a _slice_ won't error if an empty dataset is returned from the query.

```go
type Event struct {
    ID         string    `db:"id" table:"events"`
    RecordedAt time.Time `db:"recorded_at"`
}

func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgresql://...")
    if err != nil {
        panic(err)
    }
    defer conn.Close(ctx)

    var events []*Event
    err = pmx.Select(ctx, conn, &events, "select * from events limit 3")
    if err != nil {
        panic(err)
    }
    if len(events) == 0 {
        panic("no events found")
    }

    fmt.Printf("%+v\n", events)
}
```

## Update

Always provide a struct pointer.

The last argument `UpdateOptions` specifies:

- `Set`: explicit struct fields to be updated
- `By`: explicit struct fields to be matched by equality in the `where` clause

```go
type Event struct {
    ID         string    `db:"id" table:"events"`
    RecordedAt time.Time `db:"recorded_at"`
}

func main() {
    ctx := context.Background()

    conn, err := pgx.Connect(ctx, "postgresql://...")
    if err != nil {
        panic(err)
    }
    defer conn.Close(ctx)

    event := Event{
        ID:         "a1eff19b-4624-46c6-9e09-5910e7b2938d",
        RecordedAt: time.Now(),
    }

    // Generated query: update events set recorded_at = $1 where id = $2
    _, err = pmx.Update(ctx, conn, &event, &pmx.UpdateOptions{
        Set: []string{"RecordedAt"},
        By:  []string{"ID"},
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", event)
}
```

## Auto Generated Values

Given the following table with an auto generated value, e.g. serial/sequence:

```sql
create table events (
    id bigserial primary key,
    recorded_at timestamptz
);
```

Annotate the `ID` struct field with a `generated:"auto"` struct tag:

```go
type Event struct {
    ID         string    `db:"id" generated:"auto" table:"events"`
    RecordedAt time.Time `db:"recorded_at"`
}
```

The `id` column will be excluded from `insert values` and `update set` statements.

## ErrInvalidRef

The error `pmx.ErrInvalidRef` ("invalid ref") means you provided an invalid pointer reference.
