package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func main() {
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Pop out each enqueued subreddit scrape job in Redis, parse the json obj, scrape it based on its URL, then upload results to Azurite Blob
	for {
		// Pop out the next enqueued scrape job until there is none
		jobJSON, err := rdb.RPop(ctx, "scrape_jobs").Result()
		if err == redis.Nil {
			log.Println("No more jobs in queue.")
			break
		} else if err != nil {
			log.Fatalf("Failed to pop job from Redis: %v", err)
		}

		// Ensure the ScrapeJob is a valid JSON obj
		var job ScrapeJob
		if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
			log.Printf("Failed to parse job JSON: %v", err)
			continue
		}

		// Scrape the subreddit for data based on its URL, like the post titles, comments, etc.
		log.Printf("Processing job %s for %s", job.JobID, job.URL)
		result := scrapeReddit(job.JobID, job.URL)

		// Debug: print out each title on the subreddit hot page to ensure the scrape worked
		for i, post := range result.Posts {
			fmt.Printf("[%d] %s\n", i+1, post.Title)
		}

		// Upload the ScrapeResult to the Blob container raw-scrapes
		if err := uploadScrapeResult("raw-scrapes", result); err != nil {
			log.Println("Blob upload failed:", err)
		}
	}
}
