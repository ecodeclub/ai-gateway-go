package main

import (
	"fmt"
	pb "github.com/ecodeclub/ai-gateway-go/api/gen"
	GRPC "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/ecodeclub/ai-gateway-go/internal/service"
	"google.golang.org/grpc"
	"net"
	"os"
)

const (
	port = "8080"
)

var token = os.Getenv("DEEPSEEK_TOKEN")

func main() {
	svc := service.NewAIService(token)
	server := GRPC.NewServer(svc)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
	}

	s := grpc.NewServer()
	pb.RegisterAIServiceServer(s, server)
	if err := s.Serve(listener); err != nil {
		fmt.Println("", err)
	}
}
