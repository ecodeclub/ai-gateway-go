package test

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type NodeTestSuite struct {
	suite.Suite
}

func TestNode(t *testing.T) {
	suite.Run(t, new(NodeTestSuite))
}
