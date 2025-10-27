# Root Makefile for Reddit Sentiment Scraper Pipeline

.PHONY: setup all python-go orchestrator worker analysis clean

# Install everything
all: orchestrator worker analysis

# Python + Go setup
python-go: analysis orchestrator worker

# Python sentiment analysis setup
analysis:
    @echo "Installing Python dependencies for /analysis..."
    pip install -r analysis/requirements.txt

# Go orchestrator setup
orchestrator:
    @echo "Setting up Go dependencies for /orchestrator..."
    cd orchestrator && go mod tidy

# Go worker setup
worker:
    @echo "Setting up Go dependencies for /worker..."
    cd worker && go mod tidy

# Clean Go module caches (optional)
clean:
    @echo "Cleaning Go module cache..."
    go clean -modcache
