package main

import (
	"fmt"
	"os"

	"github.com/Corogura/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	st := state{
		cfg: &cfg,
	}
	cmds := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("enter command")
		os.Exit(1)
	}
	cmd := command{
		name: args[1],
		arg:  args[2:],
	}
	err = cmds.run(&st, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(len(cmd.arg))
}
