package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

// Reddit post json structure, to be incl. in ScrapeResult
type RedditPost struct {
	Title    string   `json:"title"`
	Link     string   `json:"link"`
	Comments []string `json:"comments"`
}

// Scrape result json structure, to be uploaded to Blob container
type ScrapeResult struct {
	JobID     string       `json:"job_id"`
	URL       string       `json:"url"`
	ScrapedAt string       `json:"scraped_at"`
	Posts     []RedditPost `json:"posts"`
}

// Define configurable variables from .env
var (
	userAgent   string
	maxComments int
)

func init() {
	// Load .env file from current working directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read user agents from environment
	userAgent = os.Getenv("REDDIT_USER_AGENT")
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (compatible; DevScraper/1.0)"
		log.Println("Using default user agent")
	}

	// Read max comment length (due to analysis constraints) from environemtn
	maxCommentsStr := os.Getenv("MAX_COMMENTS")
	maxComments, err = strconv.Atoi(maxCommentsStr)
	if err != nil || maxComments <= 0 {
		maxComments = 25
		log.Println("Invalid MAX_COMMENTS, defaulting to 25")
	}
}

// Function to scrape the given subreddit URL using Colly, a Go framework to easily extract structured data from websites
func scrapeReddit(jobID, url string) ScrapeResult {
	// Instantiate the json array to hold each Reddit post scraped from the front page
	posts := []RedditPost{}

	// Instantiate a collector for posts
	postCollector := colly.NewCollector(
		colly.AllowedDomains("old.reddit.com"),
		colly.UserAgent(userAgent),
	)

	// Instantiate a collector for comments
	commentCollector := colly.NewCollector(
		colly.AllowedDomains("old.reddit.com"),
		colly.UserAgent(userAgent),
	)

	// Instantiate the string array for each post that gets scraped later
	var comments []string

	// Comment collector callback (when scraper visits a Reddit post page) will only add a comment if it is more than 5 words,
	// and only up to the 25th top comment
	commentCollector.OnHTML(".comment .md", func(e *colly.HTMLElement) {
		if len(comments) >= maxComments {
			return
		}
		comment := strings.TrimSpace(e.Text)
		wordCount := len(strings.Fields(comment))

		if wordCount >= 5 {
			comments = append(comments, comment)
		}
	})

	// Add the RedditPost with the post title & its comments (filtered above) to the ScrapeResult if the title does fall within
	// ignored_keywords (usually belongs to posts that are weekly community posts that do not have relevant content to analyze)
	postCollector.OnHTML("a.title", func(e *colly.HTMLElement) {
		// Check that the title does not have any ignored_keywords
		title := e.Text
		lower := strings.ToLower(title)
		ignored_keywords := []string{"thread", "meta:", "/r/"}
		for _, keyword := range ignored_keywords {
			if strings.Contains(lower, keyword) {
				return
			}
		}

		// Extract the URL of the comment thread for the current Reddit post.
		postContainer := e.DOM.Parent().Parent()
		commentLink := postContainer.Find("ul.flat-list.buttons a.bylink.comments").AttrOr("href", "")
		if commentLink == "" {
			return
		}
		link := e.Request.AbsoluteURL(commentLink)

		// Empty the comments array to prepare for comments for this post
		comments = []string{}

		// Navigate to the comment thread URL (link) and trigger the .OnHTML callback we defined above
		if err := commentCollector.Visit(link); err != nil {
			log.Printf("Failed to visit post: %s", link)
		}

		// Add the post's title, link, and comments to the RedditPost array
		posts = append(posts, RedditPost{
			Title:    title,
			Link:     link,
			Comments: comments,
		})
	})

	// Navigate to the Reddit post URL (url) and trigger the .OnHTML callback we defined above
	err := postCollector.Visit(url)
	if err != nil {
		log.Println("Scrape error:", err)
	}

	// Return the ScrapeResult json object with the post url, RedditPost array, and metadata for auditing
	return ScrapeResult{
		JobID:     jobID,
		URL:       url,
		ScrapedAt: time.Now().UTC().Format(time.RFC3339),
		Posts:     posts,
	}
}
