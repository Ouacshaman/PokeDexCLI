package main

import (
	"fmt"
	"bufio"
	"os"
	"encoding/json"
	"io"
	"strings"
	"errors"
	"net/http"
)

type cliCommand struct {
	name        string
	description string
	callback    func(map[string]cliCommand, *listedLocation) (bool, error)
}

type listedLocation struct {
    url      string
    Count    int      `json:"count"`
    Next     string   `json:"next"`
    Previous string   `json:"previous"`
    Results  []Result `json:"results"`
}

type Result struct{
	Name string `json:"name"`
	Url string `json:"url"`
}

func main(){
	active := true
	commands := generate_cmd()
	location := &listedLocation{}
	location.url = "https://pokeapi.co/api/v2/location-area/"
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
		res, err := command.callback(commands, location)
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation) (bool, error){
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation) (bool, error){
				return false, nil
			},
		},
		"map": {
			name: "map",
			description: "list locations and explore",
			callback: func(cmds map[string]cliCommand, config *listedLocation) (bool, error){
				resp, err := http.Get(config.url)
				if err != nil{
					reply := fmt.Sprintf("unable to retrieve api: %v", err)
					fmt.Println(reply)
				}
				body, err := io.ReadAll(resp.Body)
				if resp.StatusCode > 299{
					res := fmt.Sprintf("Failure code: %s", resp.StatusCode)
					fmt.Println(res)
				}
				defer resp.Body.Close()
				err = json.Unmarshal(body, &config)
				//fmt.Printf("Unmarshaled data: %+v\n", config)
				//fmt.Println(string(body))
				if err != nil{
					reply := fmt.Sprintf("Json Body retrieval error: %v", err)
					fmt.Println(reply)
				}
				for _,location := range config.Results{
					fmt.Println(location)
				}
				config.Previous = config.url
				config.url = config.Next
				return true, nil
			},
		},
	}
}










