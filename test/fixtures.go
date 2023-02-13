package test

import (
	"time"
)

var sample = Sample{private: "used"}

type Event struct {
	Position   int64     `db:"position" generated:"auto" table:"events"`
	RecordedAt time.Time `db:"recorded_at"`
}

type Sample struct {
	ID        string `db:"id" table:"samples"`
	Transient string
	private   string
	Label     string `db:"label"`
}

type User struct {
	ID    string  `db:"id" table:"users"`
	Email string  `db:"email"`
	Token *string `db:"token"`
}
