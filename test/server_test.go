package test

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/ecodeclub/ai-gateway-go/pkg/proto"
	"google.golang.org/grpc"
	"io"
	"testing"
)

func TestServer(t *testing.T) {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Print(err)
	}
	defer conn.Close()

	client := pb.NewAIServiceClient(conn)
	stream, err := client.Ask(context.Background(), &pb.AskRequest{Text: "hello"})
	if err != nil {
		fmt.Println("方法调用失败", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
		}
		fmt.Println("接收响应", resp.GetText())
	}
}
