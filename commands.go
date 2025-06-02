package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Corogura/gator/internal/config"
	"github.com/Corogura/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	arg  []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arg) == 0 {
		return errors.New("enter username")
	}
	_, err := s.db.GetUser(context.Background(), cmd.arg[0])
	if err != nil {
		return errors.New("user does not exist")
	}
	err = s.cfg.SetUser(cmd.arg[0])
	if err != nil {
		return err
	}
	fmt.Println("user logged in successfully")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arg) < 1 {
		return errors.New("enter username")
	}
	user, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.arg[0],
		},
	)
	if err != nil {
		return err
	}
	s.cfg.SetUser(cmd.arg[0])
	fmt.Println("user registered successfully")
	fmt.Printf("user id: %s, created_at: %v, updated_at: %v, name: %s\n", user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
	return nil
}

func handlerReset(s *state, _ command) error {
	err := s.db.ResetUser(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("user reset successfully")
	return nil
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.cfg.Current_user_name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arg) < 1 {
		return errors.New("enter time period (e.g., 1h, 1d, 1w)")
	}
	fmt.Printf("Collecting feeds every %s\n", cmd.arg[0])
	timePeriod, err := time.ParseDuration(cmd.arg[0])
	if err != nil {
		return fmt.Errorf("invalid time period: %w", err)
	}
	ticker := time.NewTicker(timePeriod)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arg) < 2 {
		return errors.New("enter feed name and URL")
	}
	parsedURL, err := normURL(cmd.arg[1])
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	feed, err := s.db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.arg[0],
			Url:       parsedURL.String(),
			UserID:    user.ID,
		},
	)
	if err != nil {
		return err
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow feed: %w", err)
	}
	fmt.Printf("Feed added successfully: %s (ID: %s)\n", feed.Name, feed.ID)
	return nil
}

func handlerFeeds(s *state, _ command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user for feed %s: %w", feed.Name, err)
		}
		fmt.Printf("Name: %s, URL: %s, Username: %s\n", feed.Name, feed.Url, user.Name)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arg) < 1 {
		return errors.New("enter url to follow")
	}
	parsedURL, err := normURL(cmd.arg[0])
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	feed, err := s.db.GetFeedByURL(context.Background(), parsedURL.String())
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow feed: %w", err)
	}
	fmt.Printf("Successfully followed feed: %s by user: %s\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, _ command, user database.User) error {
	following, err := s.db.GetFeedFollowForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feeds followed by user: %w", err)
	}
	for _, follow := range following {
		fmt.Printf("Feed: %s\n", follow.FeedName)
	}
	fmt.Printf("Followed by: %s\n", user.Name)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arg) < 1 {
		return errors.New("enter feed name to unfollow")
	}
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.arg[0])
	if err != nil {
		return fmt.Errorf("failed to get feed by url: %w", err)
	}
	err = s.db.Unfollow(context.Background(), database.UnfollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow feed: %w", err)
	}
	fmt.Printf("Successfully unfollowed feed: %s\n", feed.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.arg) > 0 {
		var err error
		limit, err = strconv.Atoi(cmd.arg[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts for user: %w", err)
	}
	for _, post := range posts {
		fmt.Println("--------------------------------------------------")
		fmt.Printf("Title: %s\nURL: %s\nDescription: %s\nPublished At: %v\nFeed: %s\n",
			post.Title, post.Url, post.Description, post.PublishedAt.Time, post.FeedName)
		fmt.Println("--------------------------------------------------")
	}
	return nil
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	if _, exists := c.cmds[cmd.name]; exists {
		err := c.cmds[cmd.name](s, cmd)
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("command %s does not exist", cmd.name)
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}
