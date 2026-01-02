package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"bharani/pkg/replication"
	replicationpb "bharani/proto/replication"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "9092", "ReplicationTable server port")
	dbPath := flag.String("db", "./data/replication.db", "Database file path")
	flag.Parse()

	table, err := replication.NewTable(*dbPath)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	defer table.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	tableService := replication.NewReplicationTableService(table)
	replicationpb.RegisterReplicationTableServiceServer(s, tableService)

	log.Printf("ReplicationTable server listening on :%s", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
