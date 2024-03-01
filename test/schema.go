package test

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

var statements = []string{
	`create table events (
		position bigint default 42,
		recorded_at timestamp default make_timestamp(2024, 1, 1, 0, 0, 0),
		recorded_by text
	)`,

	`create table projections (
		id text primary key,
		name text,
		metadata jsonb,
		slice jsonb
	)`,
}

type Event struct {
	Position   int64     `db:"position"    default:"true" table:"events"`
	RecordedAt time.Time `db:"recorded_at" default:"true"`
	RecordedBy string    `db:"recorded_by"`
}

type Projection struct {
	ID                string         `db:"id" table:"projections"`
	Name              string         `db:"name"`
	Metadata          map[string]int `db:"metadata"`
	Slice             []string       `db:"slice"`
	ExportedTransient string
	privateTransient  string
}

func init() {
	_ = Projection{
		privateTransient: "suppress unused warning",
	}

	createDatabase()
	createSchema()
}

func createDatabase() {
	conn, err := pgx.Connect(context.Background(), "postgresql://user:pass@localhost/postgres")
	if err != nil {
		panic(err)
	}

	_, err = conn.Exec(context.Background(), "drop database if exists pmx")
	if err != nil {
		panic(err)
	}

	_, err = conn.Exec(context.Background(), "create database pmx")
	if err != nil {
		panic(err)
	}

	err = conn.Close(context.Background())
	if err != nil {
		panic(err)
	}
}

func createSchema() {
	conn := Connect(context.Background())
	for _, statement := range statements {
		_, err := conn.Exec(context.Background(), statement)
		if err != nil {
			panic(err)
		}
	}

	err := conn.Close(context.Background())
	if err != nil {
		panic(err)
	}
}

func Connect(ctx context.Context) *pgx.Conn {
	conn, err := pgx.Connect(ctx, "postgresql://user:pass@localhost/pmx")
	if err != nil {
		panic(err)
	}

	return conn
}
