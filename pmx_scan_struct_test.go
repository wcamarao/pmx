package pmx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/suite"
	"github.com/wcamarao/pmx"
	"github.com/wcamarao/pmx/fxt"
)

type ScanStructSuite struct {
	suite.Suite
	conn *pgx.Conn
}

func (s *ScanStructSuite) SetupTest() {
	conn, err := pgx.Connect(context.Background(), "postgresql://postgres:postgres@localhost/pmx_test")
	if err != nil {
		panic(err)
	}
	s.conn = conn
}

func TestScanStruct(t *testing.T) {
	suite.Run(t, new(ScanStructSuite))
}

func (s *ScanStructSuite) TestPointer() {
	rows, err := s.conn.Query(context.Background(), "select 'a' as id, 'b' as label")
	s.Nil(err)

	var sample fxt.Sample
	err = pmx.Scan(rows, &sample)
	s.Equal(fxt.Sample{ID: "a", Label: "b"}, sample)
	s.Nil(err)
}

func (s *ScanStructSuite) TestSkipNull() {
	rows, err := s.conn.Query(context.Background(), "select 'a' as id, null as label")
	s.Nil(err)

	var sample fxt.Sample
	err = pmx.Scan(rows, &sample)
	s.Equal(fxt.Sample{ID: "a"}, sample)
	s.Nil(err)
}

func (s *ScanStructSuite) TestSkipTransient() {
	rows, err := s.conn.Query(context.Background(), "select 'a' as id, 'b' as transient")
	s.Nil(err)

	var sample fxt.Sample
	err = pmx.Scan(rows, &sample)
	s.Equal(fxt.Sample{ID: "a"}, sample)
	s.Nil(err)
}

func (s *ScanStructSuite) TestNoRows() {
	rows, err := s.conn.Query(context.Background(), "select 1 limit 0")
	s.Nil(err)

	var sample fxt.Sample
	err = pmx.Scan(rows, &sample)
	s.Empty(sample)
	s.Nil(err)
}

func (s *ScanStructSuite) TestValue() {
	var sample fxt.Sample
	err := pmx.Scan(nil, sample)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *ScanStructSuite) TestMapPointer() {
	sample := map[string]string{}
	err := pmx.Scan(nil, &sample)
	s.Equal(pmx.ErrInvalidRef, err)
}

func (s *ScanStructSuite) TestMapValue() {
	sample := map[string]string{}
	err := pmx.Scan(nil, sample)
	s.Equal(pmx.ErrInvalidRef, err)
}
