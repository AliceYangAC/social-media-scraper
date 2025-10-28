package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/joho/godotenv"
)

var azuriteConnStr string

func init() {
	// Load .env file from current working directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read connection string from environment
	azuriteConnStr = os.Getenv("AZURITE_STORAGE_CONNECTION_STRING")
	if azuriteConnStr == "" {
		log.Fatal("AZURITE_STORAGE_CONNECTION_STRING not set in .env")
	}
}

// Function handles uploading the ScrapeResult to the Blob container for analysis; to be retired when Redis is used for queuing
func uploadScrapeResult(containerName string, result ScrapeResult) error {
	// Create context for incoming request
	ctx := context.Background()

	// Create a new Blob client from the conn string
	client, err := azblob.NewClientFromConnectionString(azuriteConnStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create blob client: %w", err)
	}

	// Validate the ScrapeResult json
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize result: %w", err)
	}

	// Set the blobname as the job ID
	blobName := fmt.Sprintf("%s.json", result.JobID)

	// Upload the blob blobName to the container containerName with the data data
	_, err = client.UploadBuffer(ctx, containerName, blobName, data, nil)
	if err != nil {
		return fmt.Errorf("failed to upload blob: %w", err)
	}

	fmt.Printf("Uploaded result to blob: %s/%s\n", containerName, blobName)
	return nil
}
