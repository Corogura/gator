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
