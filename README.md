# Reddit Social Media Scraper

## Overview

This project is a pipeline for scraping Reddit posts and comments with Go, and analyzing sentiment using both VADER and transformer-based models with Python. Sentiment analysis results are persisted to Azure Blob and Table Storage (via Azurite for local development).

## Project Architecture

```
social-media-scraper/
├── orchestrator/              # Go orchestration
│   ├── main.go
│   ├── azurite.go
│   ├── job.go
│   ├── redis.go
│   ├── go.sum
│   └── go.mod
│
├── worker/                    # Go scraping workers
│   ├── main.go
│   ├── scraper.go
│   ├── azurite.go
│   ├── job.go
│   ├── go.sum
│   └── go.mod
│
├── analysis/                 # Python sentiment analysis on Reddit titles & comments
│   ├── sentiment.py
│   ├── .env
│   └── requirements.txt
│
└── README.md
```

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/reddit-sentiment-scraper.git
cd reddit-sentiment-scraper
```

### 2. Create `.env` files for each component

Each subdirectory should contain its own `.env` file with the following configuration:

#### `/orchestrator/.env`

```env
AZURITE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;
```

#### `/worker/.env`

```env
AZURITE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;
MAX_COMMENTS=25
```

#### `/analysis/.env`

```env
AZURE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;TableEndpoint=http://127.0.0.1:10002/devstoreaccount1;
CONTAINER_NAME="raw-scrapes"
TABLE_NAME="SentimentScores"
SENTIMENT_MODEL=distilbert/distilbert-base-uncased-finetuned-sst-2-english
SENTIMENT_REVISION=714eb0f
SENTIMENT_DEVICE=-1
```

### 3. Run setup with Makefile

```bash
make           # installs everything (Go + Python)
make analysis  # installs Python dependencies from /analysis
make worker    # sets up Go modules in /worker
make orchestrator # sets up Go modules in /orchestrator
```

## 4. Start Redis and Azurite with Docker

```bash
docker run -d -p 6379:6379 redis
docker run -d -p 10000:10000 -d -p 10001:10001 -d -p 10002:10002 -d -p 10003:10003 mcr.microsoft.com/azure-storage/azurite
```

This gives you:
- Redis on `localhost:6379`
- Azurite Blob on `localhost:10000`
- Azurite Table on `localhost:10002`

---

## 5. Run the orchestrator, worker, then sentiment analysis

```bash
go run main.go redis.go azurite.go job.go
```

```bash
go run main.go scraper.go job.go azurite.go
```

```bash
py sentiment.py
```

## Components

### `/orchestrator`

- Validates subreddit input
- Triggers scraping and persistence
- Uses Go modules and `.env` for configuration

### `/worker`

- Ensures Blob and Table containers exist
- Uploads scrape results to Azurite
- Uses Go modules and `.env` for configuration

### `/analysis`

- Loads scraped JSON blobs from Azure Blob
- Applies VADER and transformer sentiment analysis
- Aggregates comment sentiment:
  - VADER: average compound score
  - Transformer: most common label across top comments
- Stores results in Azure Table Storage with:
  - `PartitionKey = job_id`
  - `RowKey = post_path`
  - Columns for sentiment scores and labels

## Viewing Results

### Table Storage

Use one of the following:

- Python: `table_client.list_entities()`
- VS Code Azurite extension
- Azure Storage Explorer GUI

### Blob Storage

Each job result is stored as `job_id.json` in the container.

## Features

- Modular Go orchestrator with Redis queueing and Azurite logging
- Configurable sentiment analysis pipeline
- Audit-friendly storage in Blob and Table
- Reproducible setup via Makefile

## Future Features

- Improve automation:
  - Redis-backed job queue: Use `scrape:jobs` and `analysis:jobs` queues to decouple orchestration, scraping, and sentiment analysis.

  - Blocking job consumption: Scraper and analysis workers (to be implemented) will use Redis `BLPop` to wait for tasks, enabling real-time responsiveness without polling overhead.

  - Job chaining: Scrape jobs will automatically enqueue analysis jobs upon completion, allowing seamless handoff between stages.
- Sentiment visualization (plotly, matplotlib, Dash):
  - Scatter Plot: Compare title sentiment vs average comment sentiment per post. Reveals agreement, disagreement, and outliers.

  - Sentiment Delta Bar Chart: Show the difference between title and comment sentiment. Highlights posts with unexpected community reactions.

  - Heatmap Matrix: Aggregate sentiment scores by subreddit. Useful for comparing tone across communities.

  - Dual-Line Time Series: Track title and comment sentiment over time. Ideal for monitoring shifts in tone or reactions to events.

  - Label Distribution Charts: Visualize transformer-based sentiment labels (POSITIVE, NEGATIVE, etc.) using stacked bars or pie charts.