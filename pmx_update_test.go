package pmx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"github.com/wcamarao/pmx"
	"github.com/wcamarao/pmx/fxt"
)

type UpdateSuite struct {
	suite.Suite
	conn *pgx.Conn
}

func (s *UpdateSuite) SetupTest() {
	conn, err := pgx.Connect(context.Background(), "postgresql://postgres:postgres@localhost/pmx_test")
	if err != nil {
		panic(err)
	}
	s.conn = conn
}

func TestUpdate(t *testing.T) {
	suite.Run(t, new(UpdateSuite))
}

func (s *UpdateSuite) TestStructPointer() {
	tag, err := s.conn.Exec(
		context.Background(),
		"insert into samples (id,label) values ($1,$2)",
		"update-pointer-id",
		"update-pointer-old-label",
	)
	s.Equal(int64(1), tag.RowsAffected())
	s.Nil(err)

	sample := fxt.Sample{
		ID:    "update-pointer-id",
		Label: "update-pointer-new-label",
	}

	err = pmx.Update(context.Background(), s.conn, &sample, []string{"Label"})
	s.Nil(err)

	var id, label string
	row := s.conn.QueryRow(context.Background(), "select * from samples where id = $1", "update-pointer-id")
	err = row.Scan(&id, &label)
	s.Equal("update-pointer-id", id)
	s.Equal("update-pointer-new-label", label)
	s.Nil(err)
}

func (s *UpdateSuite) TestSerial() {
	row := s.conn.QueryRow(
		context.Background(),
		"insert into events (label) values ($1) returning position",
		"update-serial-old-label",
	)

	var insertedPosition int64
	err := row.Scan(&insertedPosition)
	s.Nil(err)

	event := fxt.Event{
		Position: insertedPosition,
		Label:    "update-serial-new-label",
	}

	err = pmx.Update(context.Background(), s.conn, &event, []string{"Label"})
	s.Nil(err)

	var position int64
	var label string
	row = s.conn.QueryRow(context.Background(), "select * from events where label = $1", "update-serial-new-label")
	err = row.Scan(&position, &label)
	s.Equal(insertedPosition, position)
	s.Equal("update-serial-new-label", label)
	s.Nil(err)
}

func (s *UpdateSuite) TestReturning() {
	tag, err := s.conn.Exec(
		context.Background(),
		"insert into users (id,email,token) values ($1,$2,$3)",
		"update-returning-id",
		"update-returning-old-email",
		"update-returning-old-token",
	)
	s.Equal(int64(1), tag.RowsAffected())
	s.Nil(err)

	user := fxt.User{
		ID:    "update-returning-id",
		Email: "update-returning-new-email",
		Token: "update-returning-new-token",
	}

	err = pmx.Update(context.Background(), s.conn, &user, []string{"Email", "Token"})
	s.Equal("update-returning-id", user.ID)
	s.Equal("update-returning-new-email", user.Email)
	s.Equal("update-returning-new-token", user.Token)
	s.Nil(err)
}

func (s *UpdateSuite) TestAllowedFields() {
	tag, err := s.conn.Exec(
		context.Background(),
		"insert into users (id,email,token) values ($1,$2,$3)",
		"update-allowed-id",
		"update-allowed-old-email",
		"update-allowed-old-token",
	)
	s.Equal(int64(1), tag.RowsAffected())
	s.Nil(err)

	user := fxt.User{
		ID:    "update-allowed-id",
		Email: "update-allowed-new-email",
		Token: "update-allowed-new-token",
	}

	err = pmx.Update(context.Background(), s.conn, &user, []string{"Email", "Token"})
	s.Nil(err)

	var id, email, token string
	row := s.conn.QueryRow(context.Background(), "select * from users where id = $1", "update-allowed-id")
	err = row.Scan(&id, &email, &token)
	s.Equal("update-allowed-id", id)
	s.Equal("update-allowed-new-email", email)
	s.Equal("update-allowed-new-token", token)
	s.Nil(err)
}

func (s *UpdateSuite) TestUnallowedFields() {
	tag, err := s.conn.Exec(
		context.Background(),
		"insert into users (id,email,token) values ($1,$2,$3)",
		"update-unallowed-id",
		"update-unallowed-old-email",
		"update-unallowed-old-token",
	)
	s.Equal(int64(1), tag.RowsAffected())
	s.Nil(err)

	user := fxt.User{
		ID:    "update-unallowed-id",
		Email: "update-unallowed-new-email",
		Token: "update-unallowed-new-token",
	}

	err = pmx.Update(context.Background(), s.conn, &user, []string{"Email"})
	s.Nil(err)

	var id, email, token string
	row := s.conn.QueryRow(context.Background(), "select * from users where id = $1", "update-unallowed-id")
	err = row.Scan(&id, &email, &token)
	s.Equal("update-unallowed-id", id)
	s.Equal("update-unallowed-new-email", email)
	s.Equal("update-unallowed-old-token", token)
	s.Nil(err)
}

func (s *UpdateSuite) TestStructValue() {
	var sample fxt.Sample
	err := pmx.Update(context.Background(), s.conn, sample, nil)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *UpdateSuite) TestMapPointer() {
	var sample map[string]string
	err := pmx.Update(context.Background(), s.conn, &sample, nil)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *UpdateSuite) TestMapValue() {
	var sample map[string]string
	err := pmx.Update(context.Background(), s.conn, sample, nil)
	s.Equal(pmx.ErrInvalidRef, err)
}
