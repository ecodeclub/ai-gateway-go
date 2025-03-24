package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/ecodeclub/ai-gateway-go/api/ai"
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

	client := ai.NewAIServiceClient(conn)
	stream, err := client.Stream(context.Background(), &ai.StreamRequest{Text: "你好"})
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

		if resp.Final {
			fmt.Println("回答结束...")
			break
		}

		if resp.Err != "" {
			fmt.Println(resp.Err)
			break
		}
		fmt.Println(resp.Text)
	}
}
