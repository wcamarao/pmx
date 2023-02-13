package pmx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"
	"github.com/wcamarao/pmx"
	"github.com/wcamarao/pmx/test"
)

type SelectStructSuite struct {
	suite.Suite
	conn *pgx.Conn
}

func (s *SelectStructSuite) SetupTest() {
	s.conn = test.Connect()
}

func TestSelectStruct(t *testing.T) {
	suite.Run(t, new(SelectStructSuite))
}

func (s *SelectStructSuite) TestPointer() {
	var sample test.Sample
	ok, err := pmx.Select(context.Background(), s.conn, &sample, "select $1 as id, $2 as label", "a", "b")
	s.Equal(test.Sample{ID: "a", Label: "b"}, sample)
	s.Nil(err)
	s.True(ok)
}

func (s *SelectStructSuite) TestSkipNull() {
	var sample test.Sample
	ok, err := pmx.Select(context.Background(), s.conn, &sample, "select $1 as id, null as label", "a")
	s.Equal(test.Sample{ID: "a"}, sample)
	s.Nil(err)
	s.True(ok)
}

func (s *SelectStructSuite) TestSkipTransient() {
	var sample test.Sample
	ok, err := pmx.Select(context.Background(), s.conn, &sample, "select 'a' as id, 'b' as transient")
	s.Equal(test.Sample{ID: "a"}, sample)
	s.Nil(err)
	s.True(ok)
}

func (s *SelectStructSuite) TestNoRows() {
	var sample test.Sample
	ok, err := pmx.Select(context.Background(), s.conn, &sample, "select 1 limit 0")
	s.Empty(sample)
	s.Nil(err)
	s.False(ok)
}

func (s *SelectStructSuite) TestValue() {
	var sample test.Sample
	ok, err := pmx.Select(context.Background(), s.conn, sample, "select 1")
	s.Equal(pmx.ErrInvalidRef, err)
	s.False(ok)
}

func (s *SelectStructSuite) TestMapPointer() {
	sample := map[string]string{}
	ok, err := pmx.Select(context.Background(), s.conn, &sample, "select 1")
	s.Equal(pmx.ErrInvalidRef, err)
	s.False(ok)
}

func (s *SelectStructSuite) TestMapValue() {
	sample := map[string]string{}
	ok, err := pmx.Select(context.Background(), s.conn, sample, "select 1")
	s.Equal(pmx.ErrInvalidRef, err)
	s.False(ok)
}
