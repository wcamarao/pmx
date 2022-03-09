package pmx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"github.com/wcamarao/pmx"
	"github.com/wcamarao/pmx/fxt"
)

type ScanSliceSuite struct {
	suite.Suite
	conn *pgx.Conn
}

func (s *ScanSliceSuite) SetupTest() {
	conn, err := pgx.Connect(context.Background(), "postgresql://postgres:postgres@localhost/pmx_test")
	if err != nil {
		panic(err)
	}
	s.conn = conn
}

func TestScanSlice(t *testing.T) {
	suite.Run(t, new(ScanSliceSuite))
}

func (s *ScanSliceSuite) TestPointer() {
	rows, err := s.conn.Query(context.Background(), "select 'a' as id, 'b' as label")
	s.Nil(err)

	var samples []*fxt.Sample
	err = pmx.Scan(rows, &samples)
	s.Equal([]*fxt.Sample{{ID: "a", Label: "b"}}, samples)
	s.Nil(err)
}

func (s *ScanSliceSuite) TestSkipNull() {
	rows, err := s.conn.Query(context.Background(), "select 'a' as id, null as label")
	s.Nil(err)

	var samples []*fxt.Sample
	err = pmx.Scan(rows, &samples)
	s.Equal([]*fxt.Sample{{ID: "a"}}, samples)
	s.Nil(err)
}

func (s *ScanSliceSuite) TestSkipTransient() {
	rows, err := s.conn.Query(context.Background(), "select 'a' as id, 'b' as transient")
	s.Nil(err)

	var samples []*fxt.Sample
	err = pmx.Scan(rows, &samples)
	s.Equal([]*fxt.Sample{{ID: "a"}}, samples)
	s.Nil(err)
}

func (s *ScanSliceSuite) TestNoRows() {
	rows, err := s.conn.Query(context.Background(), "select 1 limit 0")
	s.Nil(err)

	var samples []*fxt.Sample
	err = pmx.Scan(rows, &samples)
	s.Empty(samples)
	s.Nil(err)
}

func (s *ScanSliceSuite) TestValue() {
	var samples []*fxt.Sample
	err := pmx.Scan(nil, samples)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *ScanSliceSuite) TestPointerOfStructValue() {
	var samples []fxt.Sample
	err := pmx.Scan(nil, &samples)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *ScanSliceSuite) TestPointerOfMapPointer() {
	var samples []*map[string]string
	err := pmx.Scan(nil, &samples)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *ScanSliceSuite) TestPointerOfMapValue() {
	var samples []map[string]string
	err := pmx.Scan(nil, &samples)
	s.Equal(pmx.ErrInvalidRef, err)
}
