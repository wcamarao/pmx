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

func (s *SelectStructSuite) SetupSuite() {
	s.conn = test.Connect(context.Background())
}

func (s *SelectStructSuite) TearDownSuite() {
	s.NoError(s.conn.Close(context.Background()))
}

func TestSelectStruct(t *testing.T) {
	suite.Run(t, new(SelectStructSuite))
}

func (s *SelectStructSuite) TestStructPointer() {
	var projection test.Projection
	err := pmx.Select(context.Background(), s.conn, &projection,
		"select $1 as id, $2 as name, $3::jsonb as metadata, $4::jsonb as slice",
		"projection-id",
		"projection-name",
		map[string]int{"index": 1},
		[]string{"value"},
	)
	s.Equal(test.Projection{
		ID:       "projection-id",
		Name:     "projection-name",
		Metadata: map[string]int{"index": 1},
		Slice:    []string{"value"},
	}, projection)
	s.NoError(err)
}

func (s *SelectStructSuite) TestNull() {
	var projection test.Projection
	err := pmx.Select(context.Background(), s.conn, &projection, "select $1 as id, $2 as name", "projection-id", nil)
	s.Equal(test.Projection{ID: "projection-id"}, projection)
	s.NoError(err)
}

func (s *SelectStructSuite) TestUnmapped() {
	var projection test.Projection
	err := pmx.Select(context.Background(), s.conn, &projection, "select $1 as id, $2 as unmapped", "projection-id", "x")
	s.Equal(test.Projection{ID: "projection-id"}, projection)
	s.NoError(err)
}

func (s *SelectStructSuite) TestNoRows() {
	var projection test.Projection
	err := pmx.Select(context.Background(), s.conn, &projection, "select 1 limit 0")
	s.ErrorIs(err, pmx.ErrNoRows)
}

func (s *SelectStructSuite) TestStructValue() {
	var projection test.Projection
	err := pmx.Select(context.Background(), s.conn, projection, "select 1")
	s.ErrorIs(err, pmx.ErrInvalidRef)
}

func (s *SelectStructSuite) TestMapPointer() {
	projection := map[string]string{}
	err := pmx.Select(context.Background(), s.conn, &projection, "select 1")
	s.ErrorIs(err, pmx.ErrInvalidRef)
}

func (s *SelectStructSuite) TestMapValue() {
	projection := map[string]string{}
	err := pmx.Select(context.Background(), s.conn, projection, "select 1")
	s.ErrorIs(err, pmx.ErrInvalidRef)
}
