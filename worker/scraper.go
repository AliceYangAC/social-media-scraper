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

type RedditPost struct {
	Title    string   `json:"title"`
	Link     string   `json:"link"`
	Comments []string `json:"comments"`
}

type ScrapeResult struct {
	JobID     string       `json:"job_id"`
	URL       string       `json:"url"`
	ScrapedAt string       `json:"scraped_at"`
	Posts     []RedditPost `json:"posts"`
}

var (
	userAgent   string
	maxComments int
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load and validate environment variables
	userAgent = os.Getenv("REDDIT_USER_AGENT")
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (compatible; DevScraper/1.0)"
		log.Println("Using default user agent")
	}

	maxCommentsStr := os.Getenv("MAX_COMMENTS")
	maxComments, err = strconv.Atoi(maxCommentsStr)
	if err != nil || maxComments <= 0 {
		maxComments = 25
		log.Println("Invalid MAX_COMMENTS, defaulting to 25")
	}
}

func scrapeReddit(jobID, url string) ScrapeResult {
	posts := []RedditPost{}

	c := colly.NewCollector(
		colly.AllowedDomains("old.reddit.com"),
		colly.UserAgent(userAgent),
	)

	commentCollector := colly.NewCollector(
		colly.AllowedDomains("old.reddit.com"),
		colly.UserAgent(userAgent),
	)

	var comments []string

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

	c.OnHTML("a.title", func(e *colly.HTMLElement) {
		title := e.Text
		lower := strings.ToLower(title)
		ignored_keywords := []string{"thread", "meta:", "/r/"}
		for _, keyword := range ignored_keywords {
			if strings.Contains(lower, keyword) {
				return
			}
		}

		postContainer := e.DOM.Parent().Parent()
		commentLink := postContainer.Find("ul.flat-list.buttons a.bylink.comments").AttrOr("href", "")
		if commentLink == "" {
			return
		}
		link := e.Request.AbsoluteURL(commentLink)
		comments = []string{} // reset for each post

		if err := commentCollector.Visit(link); err != nil {
			log.Printf("Failed to visit post: %s", link)
		}

		posts = append(posts, RedditPost{
			Title:    title,
			Link:     link,
			Comments: comments,
		})
	})

	err := c.Visit(url)
	if err != nil {
		log.Println("Scrape error:", err)
	}

	return ScrapeResult{
		JobID:     jobID,
		URL:       url,
		ScrapedAt: time.Now().UTC().Format(time.RFC3339),
		Posts:     posts,
	}
}
