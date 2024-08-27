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
	"PokeDexCLI/pokecache"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(map[string]cliCommand, *listedLocation, *pokecache.Cache) (bool, error)
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
	linkCache := pokecache.NewCache(time.Second*60) 
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
		res, err := command.callback(commands, location, linkCache)
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache) (bool, error){
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache) (bool, error){
				return false, nil
			},
		},
		"map": {
			name: "map",
			description: "list locations and explore",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache) (bool, error){
				data, found := cache.Get(config.url)
				if found{
					unmarshal(data, config)
					for _,location := range config.Results{
						fmt.Println(location.Name)
					}
					config.url = config.Next
					return true, nil
				} else{
					resp, err := http.Get(config.url)
					if err != nil{
						reply := fmt.Sprintf("unable to retrieve api: %v", err)
						fmt.Println(reply)
					}
					body, err := io.ReadAll(resp.Body)
					if err != nil{
						reply := fmt.Sprintf("Can't read response body: %v", err)
						fmt.Println(reply)
						return true, nil
					}
					if resp.StatusCode > 299{
						res := fmt.Sprintf("Failure code: %s", resp.StatusCode)
						fmt.Println(res)
						return true, nil
					}
					defer resp.Body.Close()
					cache.Add(config.url, body)
					unmarshal(body, config)
					if len(config.Next) == 0{
						fmt.Println("Reached the end, turn back")
						return true, nil
					}
					
					for _,location := range config.Results{
						fmt.Println(location.Name)
					}
					config.url = config.Next
					fmt.Println(config.Previous)
					return true, nil
				}
			},
		},
		"mapb": {
			name: "mapb",
			description: "list and go back to previous locations",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache) (bool, error){
				if len(config.Previous) == 0{
					fmt.Println("nothing in the past")
					return true, nil
				}
				resp, err := http.Get(config.Previous)
				if err != nil{
					reply := fmt.Sprintf("unable to go back: %v", err)
					fmt.Println(reply)
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil{
					reply := fmt.Sprintf("Can't read response body: %v", err)
					fmt.Println(reply)
					return true, nil
				}
				if resp.StatusCode>299{
					res := fmt.Sprintf("Failure code: %s", resp.StatusCode)
					fmt.Println(res)
					return true, nil
				}
				defer resp.Body.Close()
				unmarshal(body, config)
				for _,location := range config.Results{
					fmt.Println(location.Name)
				}
				config.url = config.Previous
				fmt.Println(config.Next)
				return true, nil
			},
		},
	}
}

func unmarshal(body []byte, config *listedLocation){
	err := json.Unmarshal(body, config)
	if err != nil{
		reply := fmt.Sprintf("Json Body retrieval error: %v", err)
		fmt.Println(reply)
	}
}



