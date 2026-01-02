package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"bharani/pkg/config"
	"bharani/pkg/volumemanager"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "9094", "VolumeManager server port")
	flag.Parse()

	cfg := config.DefaultConfig()

	_, err := volumemanager.NewManager(cfg)
	if err != nil {
		log.Fatalf("Failed to create volume manager: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	log.Printf("VolumeManager server listening on :%s", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

