package test

import (
	"context"
	testioc "github.com/ecodeclub/ai-gateway-go/internal/test/ioc"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type QuotaTestSuite struct {
	suite.Suite
	*testioc.TestApp
}

func NewQuotaSuite() *QuotaTestSuite {
	return &QuotaTestSuite{}
}

func (s *QuotaTestSuite) SetupSuite() {
	app := testioc.InitApp(testioc.TestOnly{})
	s.TestApp = app
}

func (s *QuotaTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.TestApp.DB.WithContext(ctx).Exec("TRUNCATE TABLE")
	if err != nil {
		s.T().Log(err)
	}
}

func (s *QuotaTestSuite) TestDeduct() {
	t := s.T()

	testcases := []struct {
		name string
	}{
		{
			name: "扣减临时会员表",
		},
		{
			name: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
