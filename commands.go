package main

import (
	"errors"
	"fmt"

	"github.com/Corogura/gator/internal/config"
)

type state struct {
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
	err := s.cfg.SetUser(cmd.arg[0])
	if err != nil {
		return err
	}
	fmt.Println("user set successfully")
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
