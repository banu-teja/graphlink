package db

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jDriver struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jDriver(uri, username, password string) (*Neo4jDriver, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	ctx := context.Background()
	// Verify connection
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	log.Println("Successfully connected to Neo4j")
	return &Neo4jDriver{driver: driver}, nil
}

func (n *Neo4jDriver) Close() {
	n.driver.Close(context.Background())
	log.Println("Neo4j driver closed")
}

func (n *Neo4jDriver) CreateUserNode(ctx context.Context, userID, name string) error {
	query := `
		CREATE (:User {userId: $userId, name: $name})
	`
	params := map[string]interface{}{
		"userId": userID,
		"name":   name,
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{}) // Obtain a session
	defer session.Close(ctx)                                   // Ensure session is closed

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) { // Execute in a write transaction
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("failed to create user node: %w", err)
	}
	log.Printf("User node created: %s", userID)
	return nil
}

func (n *Neo4jDriver) DeleteUserNode(ctx context.Context, userID string) error {
	query := `
		MATCH (u:User {userId: $userId})
		DETACH DELETE u
	`
	params := map[string]interface{}{
		"userId": userID,
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to delete user node: %w", err)
	}
	log.Printf("User node deleted: %s", userID)
	return nil
}

func (n *Neo4jDriver) ConnectUsers(ctx context.Context, user1ID, user2ID string) error {
	query := `
		MATCH (u1:User {userId: $user1ID}), (u2:User {userId: $user2ID})
		CREATE (u1)-[:CONNECTED_TO]->(u2)
	`
	params := map[string]interface{}{
		"user1ID": user1ID,
		"user2ID": user2ID,
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to connect users: %w", err)
	}
	log.Printf("Users connected: %s and %s", user1ID, user2ID)
	return nil
}

func (n *Neo4jDriver) DisconnectUsers(ctx context.Context, user1ID, user2ID string) error {
	query := `
		MATCH (u1:User {userId: $user1ID})-[r:CONNECTED_TO]-(u2:User {userId: $user2ID})
		DELETE r
	`
	params := map[string]interface{}{
		"user1ID": user1ID,
		"user2ID": user2ID,
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to disconnect users: %w", err)
	}
	log.Printf("Users disconnected: %s and %s", user1ID, user2ID)
	return nil
}

func (n *Neo4jDriver) GetConnectedUsers(ctx context.Context, userID string) ([]string, error) {
	query := `
		MATCH (u:User {userId: $userId})-[:CONNECTED_TO]-(connectedUser:User)
		RETURN connectedUser.userId AS connectedUserId
	`
	params := map[string]interface{}{
		"userId": userID,
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) { // Use ExecuteRead for read-only queries
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return result.Collect(ctx) // Collect all records
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get connected users: %w", err)
	}

	records, ok := result.([]*neo4j.Record) // Type assert to slice of records
	if !ok {
		return nil, fmt.Errorf("unexpected result type from ExecuteRead: %T", result)
	}

	connectedUserIDs := []string{}
	for _, record := range records {
		userIDValue, ok := record.Values[0].(string) // Access value by index
		if !ok {
			return nil, fmt.Errorf("unexpected type for connectedUserId: %T", record.Values[0])
		}
		connectedUserIDs = append(connectedUserIDs, userIDValue)
	}

	return connectedUserIDs, nil
}

func (n *Neo4jDriver) CheckConnectionPath(ctx context.Context, user1ID, user2ID string) (bool, error) {
	query := `
		MATCH (u1:User {userId: $user1ID}), (u2:User {userId: $user2ID})
		RETURN EXISTS( (u1)-[:CONNECTED_TO*1..3]-(u2) ) AS pathExists
	`
	params := map[string]interface{}{
		"user1ID": user1ID,
		"user2ID": user2ID,
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return result.Collect(ctx)
	})
	if err != nil {
		return false, fmt.Errorf("failed to execute query to check connection path: %w", err) // More specific error message
	}

	records, ok := result.([]*neo4j.Record)
	if !ok {
		return false, fmt.Errorf("unexpected result type from ExecuteRead: %T", result)
	}

	if len(records) == 0 {
		log.Printf("No records returned when checking connection path for users %s and %s. Users might not exist.", user1ID, user2ID) // Log if no records
		return false, nil                                                                                                             // Return false, indicating no path (and potentially users not found)
	}

	record := records[0]
	if len(record.Values) == 0 {
		log.Printf("No values in record when checking connection path for users %s and %s.", user1ID, user2ID) // Log if no values
		return false, nil                                                                                      // Return false, no path
	}

	pathExists, ok := record.Values[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for pathExists: %T, value: %+v", record.Values[0], record.Values[0]) // Include value in error log
	}

	return pathExists, nil
}
