package pmx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"github.com/wcamarao/pmx"
	"github.com/wcamarao/pmx/fxt"
)

type InsertSuite struct {
	suite.Suite
	conn *pgx.Conn
}

func (s *InsertSuite) SetupTest() {
	conn, err := pgx.Connect(context.Background(), "postgresql://postgres:postgres@localhost/pmx_test")
	if err != nil {
		panic(err)
	}
	s.conn = conn
}

func TestInsert(t *testing.T) {
	suite.Run(t, new(InsertSuite))
}

func (s *InsertSuite) TestStructPointer() {
	sample := fxt.Sample{
		ID:    "insert-pointer-id",
		Label: "insert-pointer-label",
	}

	err := pmx.Insert(context.Background(), s.conn, &sample)
	s.Nil(err)

	var id, label string
	row := s.conn.QueryRow(context.Background(), "select * from samples where id = $1", "insert-pointer-id")
	err = row.Scan(&id, &label)
	s.Equal("insert-pointer-id", id)
	s.Equal("insert-pointer-label", label)
	s.Nil(err)
}

func (s *InsertSuite) TestSerial() {
	event := fxt.Event{
		Label: "insert-serial-label",
	}

	err := pmx.Insert(context.Background(), s.conn, &event)
	s.Nil(err)

	var position int
	var label string
	row := s.conn.QueryRow(context.Background(), "select * from events where label = $1", "insert-serial-label")
	err = row.Scan(&position, &label)
	s.Equal("insert-serial-label", label)
	s.NotZero(position)
	s.Nil(err)
}

func (s *InsertSuite) TestReturning() {
	event := fxt.Event{
		Label: "insert-returning-label",
	}

	err := pmx.Insert(context.Background(), s.conn, &event)
	s.Equal("insert-returning-label", event.Label)
	s.NotZero(event.Position)
	s.Nil(err)
}

func (s *InsertSuite) TestStructValue() {
	var sample fxt.Sample
	err := pmx.Insert(context.Background(), s.conn, sample)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *InsertSuite) TestMapPointer() {
	var sample map[string]string
	err := pmx.Insert(context.Background(), s.conn, &sample)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *InsertSuite) TestMapValue() {
	var sample map[string]string
	err := pmx.Insert(context.Background(), s.conn, sample)
	s.Equal(pmx.ErrInvalidRef, err)
}
