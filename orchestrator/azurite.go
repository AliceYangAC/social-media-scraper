package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/joho/godotenv"
)

var azuriteConnStr string

func init() {
	// Load .env file from current working directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Error loading .env file")
	}

	// Read connection string from environment
	azuriteConnStr = os.Getenv("AZURITE_STORAGE_CONNECTION_STRING")
	if azuriteConnStr == "" {
		log.Fatal("❌ AZURITE_STORAGE_CONNECTION_STRING not set in .env")
	}
}

func ensureBlobContainer(containerName string) {
	ctx := context.Background()

	client, err := azblob.NewClientFromConnectionString(azuriteConnStr, nil)
	if err != nil {
		log.Fatalf("Failed to create Blob client: %v", err)
	}

	containerClient := client.ServiceClient().NewContainerClient(containerName)

	_, err = containerClient.Create(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == "ContainerAlreadyExists" {
			fmt.Printf("Blob container '%s' already exists.\n", containerName)
		} else {
			log.Fatalf("Failed to create container '%s': %v", containerName, err)
		}
	} else {
		fmt.Printf("Created blob container '%s'.\n", containerName)
	}
}

func ensureTable(tableName string) {
	ctx := context.Background()

	serviceClient, err := aztables.NewServiceClientFromConnectionString(azuriteConnStr, nil)
	if err != nil {
		log.Fatalf("Failed to create Table service client: %v", err)
	}

	_, err = serviceClient.CreateTable(ctx, tableName, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == "TableAlreadyExists" {
			log.Printf("Table '%s' already exists.\n", tableName)
		} else {
			log.Fatalf("Failed to create table '%s': %v", tableName, err)
		}
	} else {
		log.Printf("Created table '%s'.\n", tableName)
	}
}
