package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

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
