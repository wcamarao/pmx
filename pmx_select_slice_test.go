package pmx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"
	"github.com/wcamarao/pmx"
	"github.com/wcamarao/pmx/test"
)

type SelectSliceSuite struct {
	suite.Suite
	conn *pgx.Conn
}

func (s *SelectSliceSuite) SetupSuite() {
	s.conn = test.Connect(context.Background())
}

func (s *SelectSliceSuite) TearDownSuite() {
	s.NoError(s.conn.Close(context.Background()))
}

func TestSelectSlice(t *testing.T) {
	suite.Run(t, new(SelectSliceSuite))
}

func (s *SelectSliceSuite) TestSlicePointer() {
	var projections []*test.Projection
	err := pmx.Select(context.Background(), s.conn, &projections,
		"select $1 as id, $2 as name, $3::jsonb as metadata, $4::jsonb as slice",
		"projection-id",
		"projection-name",
		map[string]int{"index": 1},
		[]string{"value"},
	)
	s.Equal([]*test.Projection{{
		ID:       "projection-id",
		Name:     "projection-name",
		Metadata: map[string]int{"index": 1},
		Slice:    []string{"value"},
	}}, projections)
	s.NoError(err)
}

func (s *SelectSliceSuite) TestNull() {
	var projections []*test.Projection
	err := pmx.Select(context.Background(), s.conn, &projections, "select $1 as id, $2 as name", "projection-id", nil)
	s.Equal([]*test.Projection{{ID: "projection-id"}}, projections)
	s.NoError(err)
}

func (s *SelectSliceSuite) TestUnmapped() {
	var projections []*test.Projection
	err := pmx.Select(context.Background(), s.conn, &projections, "select $1 as id, $2 as unmapped", "projection-id", "x")
	s.Equal([]*test.Projection{{ID: "projection-id"}}, projections)
	s.NoError(err)
}

func (s *SelectSliceSuite) TestNoRows() {
	var projections []*test.Projection
	err := pmx.Select(context.Background(), s.conn, &projections, "select 1 limit 0")
	s.Empty(projections)
	s.NoError(err)
}

func (s *SelectSliceSuite) TestSliceValue() {
	var projections []*test.Projection
	err := pmx.Select(context.Background(), s.conn, projections, "select 1")
	s.ErrorIs(err, pmx.ErrInvalidRef)
}

func (s *SelectSliceSuite) TestSlicePointerOfStructValue() {
	var projections []test.Projection
	err := pmx.Select(context.Background(), s.conn, &projections, "select 1")
	s.ErrorIs(err, pmx.ErrInvalidRef)
}

func (s *SelectSliceSuite) TestSlicePointerOfMapPointer() {
	var projections []*map[string]string
	err := pmx.Select(context.Background(), s.conn, &projections, "select 1")
	s.ErrorIs(err, pmx.ErrInvalidRef)
}

func (s *SelectSliceSuite) TestSlicePointerOfMapValue() {
	var projections []map[string]string
	err := pmx.Select(context.Background(), s.conn, &projections, "select 1")
	s.ErrorIs(err, pmx.ErrInvalidRef)
}
