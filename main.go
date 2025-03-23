package main

import (
	"fmt"
	GRPC "github.com/ecodeclub/ai-gateway-go/grpc"
	pb "github.com/ecodeclub/ai-gateway-go/pkg/proto"
	"google.golang.org/grpc"
	"net"
)

const (
	port = "8080"
)

func main() {
	server := &GRPC.Server{}
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
