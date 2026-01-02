package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"bharani/pkg/config"
	"bharani/pkg/master"
	masterpb "bharani/proto/master"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "9093", "Master server port")
	cellID := flag.String("cell", "cell1", "Cell ID")
	replicationAddr := flag.String("replication", "localhost:9092", "ReplicationTable address")
	flag.Parse()

	cfg := config.DefaultConfig()
	cfg.CellID = *cellID

	masterInstance, err := master.NewMaster(cfg, *cellID, *replicationAddr)
	if err != nil {
		log.Fatalf("Failed to create master: %v", err)
	}
	defer masterInstance.Close()

	ctx := context.Background()
	go masterInstance.MonitorOSDs(ctx)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	masterService := master.NewMasterService(masterInstance)
	masterpb.RegisterMasterServiceServer(s, masterService)

	log.Printf("Master server listening on :%s", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

