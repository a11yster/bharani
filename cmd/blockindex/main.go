package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"bharani/pkg/blockindex"
	blockindexpb "bharani/proto/blockindex"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "9091", "BlockIndex server port")
	dbPath := flag.String("db", "./data/blockindex.db", "Database file path")
	flag.Parse()

	index, err := blockindex.NewIndex(*dbPath)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	defer index.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	indexService := blockindex.NewBlockIndexService(index)
	blockindexpb.RegisterBlockIndexServiceServer(s, indexService)

	log.Printf("BlockIndex server listening on :%s", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
