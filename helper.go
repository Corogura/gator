package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Corogura/gator/internal/database"
)

func normURL(u string) (*url.URL, error) {
	// Ensure the URL has a scheme
	if !strings.Contains(u, "://") {
		u = "https://" + u
	}
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure the URL has a host
	if parsedURL.Host == "" {
		return nil, errors.New("URL must contain a host")
	}

	return parsedURL, nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.Current_user_name == "" {
			return errors.New("user not logged in")
		}
		user, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func parsePubDate(pubDate string) (time.Time, error) {
	// Parse the publication date from the RSS item
	pubTime, err := time.Parse(time.RFC1123, pubDate)
	if err == nil {
		return pubTime, nil
	}
	// If RFC1123 fails, try RFC3339
	pubTime, err = time.Parse(time.RFC3339, pubDate)
	if err == nil {
		return pubTime, nil
	}
	// If both formats fail, try RFC822
	pubTime, err = time.Parse(time.RFC822, pubDate)
	if err == nil {
		return pubTime, nil
	}
	return time.Time{}, fmt.Errorf("failed to parse publication date: %w", err)
}
