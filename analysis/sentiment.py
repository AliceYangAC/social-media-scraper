from azure.storage.blob import BlobServiceClient
import json
import pandas as pd
from nltk.sentiment import SentimentIntensityAnalyzer
from transformers import pipeline
from urllib.parse import urlparse
import os
from dotenv import load_dotenv
from azure.data.tables import TableServiceClient, UpdateMode

# Get the directory of the current script
script_dir = os.path.dirname(os.path.abspath(__file__))

# Build the full path to the .env file
env_path = os.path.join(script_dir, ".env")

# Load the .env file
load_dotenv(dotenv_path=env_path)

# import nltk
# nltk.download("vader_lexicon")

CONN_STR = os.getenv("AZURE_STORAGE_CONNECTION_STRING")
CONTAINER_NAME = os.getenv("CONTAINER_NAME")
MODEL_NAME = os.getenv("SENTIMENT_MODEL", "distilbert/distilbert-base-uncased-finetuned-sst-2-english")
REVISION = os.getenv("SENTIMENT_REVISION", "714eb0f")
DEVICE = int(os.getenv("SENTIMENT_DEVICE", "-1"))
TABLE_NAME = os.getenv("TABLE_NAME", "SentimentScores")
table_service = TableServiceClient.from_connection_string(conn_str=CONN_STR)
table_client = table_service.get_table_client(TABLE_NAME)
blob_service = BlobServiceClient.from_connection_string(CONN_STR)
container_client = blob_service.get_container_client(CONTAINER_NAME)
sentiment_pipeline = pipeline("sentiment-analysis", model=MODEL_NAME, revision=REVISION, device=DEVICE)

# Ensure table exists
try:
    table_client.create_table()
except:
    pass  # already exists

# List all blobs and process each one
for blob in container_client.list_blobs():
    blob_name = blob.name
    blob_client = blob_service.get_blob_client(container=CONTAINER_NAME, blob=blob_name)
    data = json.loads(blob_client.download_blob().readall())

    job_id = data.get("job_id", blob_name.replace(".json", ""))
    scraped_at = data.get("scraped_at", "unknown")
    url = data.get("url", "unknown")

    titles = [post["title"] for post in data.get("posts", [])]
    comments_list = [post.get("comments", []) for post in data.get("posts", [])]
    post_list = [urlparse(post.get("link", [])).path for post in data.get("posts", [])]

    df = pd.DataFrame({
        "title": titles,
        "comments": comments_list,
        "post": post_list
    })
    vader = SentimentIntensityAnalyzer()

    df["vader_title"] = df["title"].apply(lambda t: vader.polarity_scores(t)["compound"])

    # Average sentiment across comments
    def vader_comment_score(comments):
        if not comments:
            return None
        scores = [vader.polarity_scores(c)["compound"] for c in comments]
        return sum(scores) / len(scores)

    df["vader_comments_avg"] = df["comments"].apply(vader_comment_score)
    
    # Transformer sentiment using configured pipeline
    df["transformer_title"] = df["title"].apply(lambda t: sentiment_pipeline(t)[0]["label"])

    def truncate(text, max_words=100):
        return " ".join(text.split()[:max_words])

    def transformer_comment_score(comments):
        if not comments:
            return None
        truncated = [truncate(c) for c in comments[:25]]  # limit to top 25 comments
        results = sentiment_pipeline(truncated)
        labels = [r["label"] for r in results]
        return max(set(labels), key=labels.count)  # most common label

    df["transformer_comments_mode"] = df["comments"].apply(transformer_comment_score)
    
    for _, row in df.iterrows():
        entity = {
            "PartitionKey": job_id,
            "RowKey": row["post"].replace("/", "_"),  # sanitize path
            "VaderTitle": row["vader_title"],
            "VaderCommentsAvg": row["vader_comments_avg"],
            "TransformerTitle": row["transformer_title"],
            "TransformerCommentsMode": row["transformer_comments_mode"],
            "ScrapedAt": scraped_at,
            "SourceURL": url
        }
    table_client.upsert_entity(entity=entity, mode=UpdateMode.REPLACE)

    
    print(df[["post", "vader_title", "vader_comments_avg", "transformer_title", "transformer_comments_mode"]].head())

