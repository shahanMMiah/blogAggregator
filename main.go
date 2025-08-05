package main

import (
	"fmt"
	"os"

	"github.com/shahanmmiah/blogAggregator/internal/config"
)

func main() {

	c, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}

	state := config.State{Config_ptr: &c}
	cmds := config.Commands{Cmds: map[string]func(*config.State, config.Command) error{}}

	cmds.Register("login", config.HandlerLogin)

	inputArgs := os.Args
	if len(inputArgs) < 2 {
		fmt.Printf("Error: No command argument specified")
		os.Exit(1)
	}

	cmd := config.Command{Name: inputArgs[1], Args: inputArgs[2:]}
	err = cmds.Run(&state, cmd)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
