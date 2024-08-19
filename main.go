package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"errors"
)

type cliCommand struct {
	name        string
	description string
	callback    func(map[string]cliCommand) (bool, error)
}

func main(){
	active := true
	commands := generate_cmd()
	for active{
		fmt.Printf("pokedex > ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err := scanner.Err()
		if err != nil{
			break
		}
		argument := strings.ToLower(scanner.Text())
		command, ok := commands[argument]
		if !ok{
			err := errors.New("invalid command")
			fmt.Println(err)
			continue
		}
		res, err := command.callback(commands)
		if err != nil{
			fmt.Println(err)
		}
		active = res
	}
}

func generate_cmd() map[string]cliCommand{
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    func(cmds map[string]cliCommand) (bool, error){
				fmt.Println("Welcome to the Pokedex:")
				fmt.Println("Usage:")
				for _,cmd := range cmds{
					fmt.Println(cmd.name, ": ", cmd.description)
				}
				return true, nil
			},
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    func(cmds map[string]cliCommand) (bool, error){
				return false, nil
			},
		},
	}
}










