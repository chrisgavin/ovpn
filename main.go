package main

import (
	"github.com/chrisgavin/ovpn/cmd"
)

func main() {
	rootCommand, err := cmd.NewRootCommand()
	if err != nil {
		panic(err)
	}
	rootCommand.Run()
}
