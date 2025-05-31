package main

import (
	"context"
	"errors"
	"fmt"
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

func handlerAgg(s *state, _ command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}
	fmt.Printf("Title: %s\n", feed.Channel.Title)
	fmt.Printf("Link: %s\n", feed.Channel.Link)
	fmt.Printf("Description: %s\n", feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		fmt.Printf("Item[%d] Title: %s\n", i, item.Title)
		fmt.Printf("Item[%d] Link: %s\n", i, item.Link)
		fmt.Printf("Item[%d] Description: %s\n", i, item.Description)
		fmt.Printf("Item[%d] PubDate: %s\n", i, item.PubDate)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arg) < 2 {
		return errors.New("enter feed name and URL")
	}
	currentUser, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	feed, err := s.db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.arg[0],
			Url:       cmd.arg[1],
			UserID:    currentUser.ID,
		},
	)
	if err != nil {
		return err
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
