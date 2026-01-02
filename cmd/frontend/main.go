package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"bharani/pkg/config"
	"bharani/pkg/frontend"
	frontendpb "bharani/proto/frontend"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "8080", "Frontend server port")
	blockIndexAddr := flag.String("blockindex", "localhost:9091", "BlockIndex address")
	replicationAddr := flag.String("replication", "localhost:9092", "ReplicationTable address")
	masterAddr := flag.String("master", "localhost:9093", "Master address")
	flag.Parse()

	cfg := config.DefaultConfig()

	frontendInstance, err := frontend.NewFrontend(cfg, *blockIndexAddr, *replicationAddr, *masterAddr)
	if err != nil {
		log.Fatalf("Failed to create frontend: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	frontendService := frontend.NewFrontendService(frontendInstance)
	frontendpb.RegisterFrontendServiceServer(s, frontendService)

	log.Printf("Frontend server listening on :%s", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

