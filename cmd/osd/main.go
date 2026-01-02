package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"bharani/pkg/config"
	"bharani/pkg/osd"
	osdpb "bharani/proto/osd"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "9090", "OSD server port")
	address := flag.String("address", "localhost:9090", "OSD address")
	cellID := flag.String("cell", "cell1", "Cell ID")
	dataDir := flag.String("data-dir", "./data/osd", "Data directory for blocks")
	flag.Parse()

	cfg := config.DefaultConfig()
	cfg.OSDPort = *port
	cfg.OSDDataDir = *dataDir
	cfg.CellID = *cellID

	osdInstance, err := osd.NewOSD(cfg, *address, *cellID)
	if err != nil {
		log.Fatalf("Failed to create OSD: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	osdService := osd.NewOSDService(osdInstance)
	osdpb.RegisterOSDServiceServer(s, osdService)

	log.Printf("OSD server listening on :%s", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
