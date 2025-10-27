package main

// Scrape job json structure
type ScrapeJob struct {
	URL       string `json:"url"`
	JobID     string `json:"job_id"`
	CreatedAt string `json:"created_at"`
	Priority  int    `json:"priority"`
	Notes     string `json:"notes"`
}
