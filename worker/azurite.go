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

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func uploadScrapeResult(containerName string, result ScrapeResult) error {
	ctx := context.Background()

	connStr := os.Getenv("AZURITE_STORAGE_CONNECTION_STRING")
	if connStr == "" {
		return fmt.Errorf("AZURITE_STORAGE_CONNECTION_STRING not set in environment")
	}

	client, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create blob client: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize result: %w", err)
	}

	blobName := fmt.Sprintf("%s.json", result.JobID)
	_, err = client.UploadBuffer(ctx, containerName, blobName, data, nil)
	if err != nil {
		return fmt.Errorf("failed to upload blob: %w", err)
	}

	fmt.Printf("✅ Uploaded result to blob: %s/%s\n", containerName, blobName)
	return nil
}
