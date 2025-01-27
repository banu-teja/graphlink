package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/banu-teja/graphlink/internal/config"
	"github.com/banu-teja/graphlink/internal/db"
	"github.com/banu-teja/graphlink/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/banu-teja/graphlink/pkg/api/graph"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	neo4jDriver, err := db.NewNeo4jDriver(cfg.Neo4jURI, cfg.Neo4jUsername, cfg.Neo4jPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer neo4jDriver.Close()

	graphService := service.NewGraphService(neo4jDriver)
	grpcServer := grpc.NewServer()
	pb.RegisterGraphServiceServer(grpcServer, graphService)

	reflection.Register(grpcServer) // Register reflection service on gRPC server

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on port: %d", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
