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
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	for {
		jobJSON, err := rdb.RPop(ctx, "scrape_jobs").Result()
		if err == redis.Nil {
			log.Println("✅ No more jobs in queue.")
			break
		} else if err != nil {
			log.Fatalf("❌ Failed to pop job from Redis: %v", err)
		}

		var job ScrapeJob
		if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
			log.Printf("⚠️ Failed to parse job JSON: %v", err)
			continue
		}

		log.Printf("🔍 Processing job %s for %s", job.JobID, job.URL)
		result := scrapeReddit(job.JobID, job.URL)

		for i, post := range result.Posts {
			fmt.Printf("[%d] %s\n", i+1, post.Title)
		}

		if err := uploadScrapeResult("raw-scrapes", result); err != nil {
			log.Println("⚠️ Blob upload failed:", err)
		}
	}
}
