package test

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var schema = []string{
	`create table events (
		position bigserial,
		recorded_at timestamptz
	)`,

	`create table samples (
		id text primary key,
		label text
	)`,

	`create table users (
		id text,
		email text,
		token text
	)`,
}

func init() {
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

	conn = Connect()
	for _, sql := range schema {
		_, err = conn.Exec(context.Background(), sql)
		if err != nil {
			panic(err)
		}
	}
}

func Connect() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), "postgresql://user:pass@localhost/pmx")
	if err != nil {
		panic(err)
	}

	return conn
}
