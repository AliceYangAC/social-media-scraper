package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func enqueueJob(rdb *redis.Client, job ScrapeJob) error {
	// Ensure the scrape job is a valid json obj
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// If valid json, push the job to the rdb Redis client
	err = rdb.LPush(ctx, "scrape_jobs", jobJSON).Err()
	if err == nil {
		fmt.Printf("Enqueued job %s for %s\n", job.JobID, job.URL)
	}
	return err
}
