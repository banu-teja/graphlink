package service

import (
	"context"
	"fmt"
	"log"

	"github.com/banu-teja/graphlink/internal/db"
	pb "github.com/banu-teja/graphlink/pkg/api/graph"
)

type GraphService struct {
	pb.UnimplementedGraphServiceServer
	neo4jDriver *db.Neo4jDriver
}

func NewGraphService(neo4jDriver *db.Neo4jDriver) *GraphService {
	return &GraphService{neo4jDriver: neo4jDriver}
}

func (s *GraphService) CreateUserNode(ctx context.Context, req *pb.CreateUserNodeRequest) (*pb.CreateUserNodeResponse, error) {
	if err := s.neo4jDriver.CreateUserNode(ctx, req.GetUserId(), req.GetName()); err != nil {
		log.Printf("Failed to create user node: %v", err)
		return nil, fmt.Errorf("failed to create user node")
	}
	return &pb.CreateUserNodeResponse{Success: true}, nil
}

func (s *GraphService) DeleteUserNode(ctx context.Context, req *pb.DeleteUserNodeRequest) (*pb.DeleteUserNodeResponse, error) {
	if err := s.neo4jDriver.DeleteUserNode(ctx, req.GetUserId()); err != nil {
		log.Printf("Failed to delete user node: %v", err)
		return nil, fmt.Errorf("failed to delete user node")
	}
	return &pb.DeleteUserNodeResponse{Success: true}, nil
}

func (s *GraphService) ConnectUsers(ctx context.Context, req *pb.ConnectUsersRequest) (*pb.ConnectUsersResponse, error) {
	if err := s.neo4jDriver.ConnectUsers(ctx, req.GetUserId_1(), req.GetUserId_2()); err != nil {
		log.Printf("Failed to connect users: %v", err)
		return nil, fmt.Errorf("failed to connect users")
	}
	return &pb.ConnectUsersResponse{Success: true}, nil
}

func (s *GraphService) DisconnectUsers(ctx context.Context, req *pb.DisconnectUsersRequest) (*pb.DisconnectUsersResponse, error) {
	if err := s.neo4jDriver.DisconnectUsers(ctx, req.GetUserId_1(), req.GetUserId_2()); err != nil {
		log.Printf("Failed to disconnect users: %v", err)
		return nil, fmt.Errorf("failed to disconnect users")
	}
	return &pb.DisconnectUsersResponse{Success: true}, nil
}

func (s *GraphService) GetConnectedUsers(ctx context.Context, req *pb.GetConnectedUsersRequest) (*pb.GetConnectedUsersResponse, error) {
	connectedUserIDs, err := s.neo4jDriver.GetConnectedUsers(ctx, req.GetUserId())
	if err != nil {
		log.Printf("Failed to get connected users: %v", err)
		return nil, fmt.Errorf("failed to get connected users")
	}
	return &pb.GetConnectedUsersResponse{ConnectedUserIds: connectedUserIDs}, nil
}

func (s *GraphService) CheckConnectionPath(ctx context.Context, req *pb.CheckConnectionPathRequest) (*pb.CheckConnectionPathResponse, error) {
	pathExists, err := s.neo4jDriver.CheckConnectionPath(ctx, req.GetUserId_1(), req.GetUserId_2())
	if err != nil {
		log.Printf("Failed to check connection path: %v", err)
		return nil, fmt.Errorf("failed to check connection path")
	}
	return &pb.CheckConnectionPathResponse{PathExists: pathExists}, nil
}
