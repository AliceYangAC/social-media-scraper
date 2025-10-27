package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func enqueueJob(rdb *redis.Client, job ScrapeJob) error {
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return err
	}

	err = rdb.LPush(ctx, "scrape_jobs", jobJSON).Err()
	if err == nil {
		fmt.Printf("Enqueued job %s for %s\n", job.JobID, job.URL)
	}
	return err
}
