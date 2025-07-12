package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MongoURI       = "mongodb+srv://thrillseekernw:1115@cluster0.zsstj4i.mongodb.net"
	DatabaseName   = "go_test"
	CollectionName = "api_keys"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(DatabaseName)
	collection := db.Collection(CollectionName)

	// Create unique index on api_key field
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "api_key", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: failed to create index: %v", err)
	}

	// Sample API keys for testing
	apiKeys := []string{
		"test-api-key-1",
		"test-api-key-2",
		"demo-api-key-123",
	}

	// Insert API keys
	for _, apiKey := range apiKeys {
		_, err := collection.InsertOne(ctx, bson.M{
			"api_key":   apiKey,
			"created_at": time.Now(),
			"active":     true,
		})
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				log.Printf("API key %s already exists", apiKey)
			} else {
				log.Printf("Failed to insert API key %s: %v", apiKey, err)
			}
		} else {
			log.Printf("Successfully added API key: %s", apiKey)
		}
	}

	// List all API keys
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to find API keys: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatalf("Failed to decode results: %v", err)
	}

	fmt.Println("\nCurrent API keys in database:")
	for _, result := range results {
		fmt.Printf("- %s\n", result["api_key"])
	}

	fmt.Println("\nSetup complete! You can now use these API keys to test the API.")
} 