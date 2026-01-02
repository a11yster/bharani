package main

import (
	"context"
	"fmt"
	"log"

	"bharani/proto/frontend"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := frontend.NewFrontendServiceClient(conn)
	ctx := context.Background()

	putData := []byte("Hello, Magic Pocket!")
	putResp, err := client.Put(ctx, &frontend.PutRequest{
		Data: putData,
	})
	if err != nil {
		log.Fatalf("Put failed: %v", err)
	}

	if !putResp.Success {
		log.Fatalf("Put failed: %s", putResp.Error)
	}

	fmt.Printf("Put successful! Hash: %s\n", putResp.Hash)

	getResp, err := client.Get(ctx, &frontend.GetRequest{
		Hash: putResp.Hash,
	})
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}

	if !getResp.Success {
		log.Fatalf("Get failed: %s", getResp.Error)
	}

	fmt.Printf("Get successful! Data: %s\n", string(getResp.Data))
}
