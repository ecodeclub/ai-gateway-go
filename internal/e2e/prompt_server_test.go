package e2e

import (
	promptv1 "github.com/ecodeclub/ai-gateway-go/api/gen/prompt/v1"
	grpc2 "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/repository"
	dao2 "github.com/ecodeclub/ai-gateway-go/internal/repository/dao"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"testing"
)

func TestPromptServer(t *testing.T) {
	server := grpc.NewServer()
	defer server.GracefulStop()

	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13316)/ai_gateway?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s"))
	require.NoError(t, err)
	dao := dao2.NewPromptDAO(db)
	repo := repository.NewPromptRepo(dao)
	svc := service.NewPromptService(repo)
	promptServer := grpc2.NewPromptServer(svc)
	promptv1.RegisterPromptServiceServer(server, promptServer)
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	err = server.Serve(l)
	require.NoError(t, err)
}
