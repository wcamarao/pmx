package pmx_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wcamarao/pmx"
	"github.com/wcamarao/pmx/fxt"
)

type IsZeroSuite struct {
	suite.Suite
}

func TestIsZero(t *testing.T) {
	suite.Run(t, new(IsZeroSuite))
}

func (s *IsZeroSuite) TestPointers() {
	s.True(pmx.IsZero(&fxt.Sample{}))
	s.False(pmx.IsZero(&fxt.Sample{ID: "a"}))
}

func (s *IsZeroSuite) TestValues() {
	s.True(pmx.IsZero(fxt.Sample{}))
	s.False(pmx.IsZero(fxt.Sample{ID: "a"}))
}

func (s *IsZeroSuite) TestNil() {
	s.True(pmx.IsZero(nil))
}
