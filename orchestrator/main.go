package main

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// var ctx = context.Background()

func main() {
	// Ensure Azurite containers and tables exist
	ensureBlobContainer("raw-scrapes")
	ensureTable("ScrapeJobs")

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Enqueue jobs
	urls := []string{
		"https://old.reddit.com/r/canada/",
		"https://old.reddit.com/r/worldnews/",
		"https://old.reddit.com/r/technology/",
	}

	for _, url := range urls {
		job := ScrapeJob{
			URL:       url,
			JobID:     uuid.New().String(),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			Priority:  1,
			Notes:     "Initial test scrape",
		}

		if err := enqueueJob(rdb, job); err != nil {
			log.Println("Failed to enqueue job:", err)
		}
	}
}
