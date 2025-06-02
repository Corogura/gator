package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/Corogura/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch feed: " + resp.Status)
	}

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(dat, &feed); err != nil {
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}
	return &feed, nil
}

func scrapeFeeds(s *state) error {
	feeds, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get next feed to fetch: %w", err)
	}
	feed, err := s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID:        feeds.ID,
		UpdatedAt: time.Now(),
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fetchedFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}
	fmt.Printf("Fetched feed: %s\n", fetchedFeed.Channel.Title)
	for _, item := range fetchedFeed.Channel.Item {
		pubDate, err := parsePubDate(item.PubDate)
		var parsed sql.NullTime
		if err != nil {
			parsed = sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			}
		} else {
			parsed = sql.NullTime{
				Time:  pubDate,
				Valid: true,
			}
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			FeedID:      feed.ID,
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: parsed,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				continue
			} else {
				return fmt.Errorf("failed to create post: %w", err)
			}
		}
	}
	return nil
}
