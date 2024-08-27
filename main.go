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
	callback    func(map[string]cliCommand, *listedLocation, *pokecache.Cache, []string) (bool, error)
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

type Explore struct{
	Pokemon_encounters []Pokemon_encounters `json:"pokemon_encounters"`
}

type Pokemon_encounters struct{
	Pokemon Pokemon `json:"pokemon"`
}

type Pokemon struct{
	Name string `json:"name"`
}

func main(){
	active := true
	commands := generate_cmd()
	location := &listedLocation{}
	location.url = "https://pokeapi.co/api/v2/location-area/"
	linkCache := pokecache.NewCache(time.Second*60)
	scanner := bufio.NewScanner(os.Stdin)
	for active{
		fmt.Printf("pokedex > ")
		scanner.Scan()
		err := scanner.Err()
		if err != nil{
			break
		}
		argument := strings.ToLower(scanner.Text())
		args := strings.Split(argument, " ")
		command, ok := commands[args[0]]
		params := args[1:]
		if !ok{
			err := errors.New("invalid command")
			fmt.Println(err)
			continue
		}
		res, err := command.callback(commands, location, linkCache, params)
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string) (bool, error){
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string) (bool, error){
				return false, nil
			},
		},
		"map": {
			name: "map",
			description: "list locations and explore",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string) (bool, error){
				data, found := cache.Get(config.url)
				if found{
					unmarshal(data, config)
					location_print(config.Results)
				} else{
					body, error := httpGet(config.url)
					if error{
						return true, nil
					}
					cache.Add(config.url, body)
					//fmt.Println(string(body))
					unmarshal(body, config)
					if len(config.Next) == 0{
						fmt.Println("Reached the end, turn back")
						return true, nil
					}
					location_print(config.Results)
				}
				config.url = config.Next
				return true, nil
			},
		},
		"mapb": {
			name: "mapb",
			description: "list and go back to previous locations",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string) (bool, error){
				data, found := cache.Get(config.Previous)
				if found{
					unmarshal(data, config)
					location_print(config.Results)
				} else{
					if len(config.Previous) == 0 || config.Previous == ""{
						fmt.Println("nothing in the past")
						return true, nil
					}
					body, error := httpGet(config.Previous)
					if error{
						return true, nil
					}
					cache.Add(config.Previous, body)
					result := unmarshal(body, config)
					if !result{
						return true, nil
					}
					location_print(config.Results)
				}
				config.url = config.Previous
				return true, nil
			},
		},
		"explore": {
			name: "explore",
			description: "explore and list pokemons",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string) (bool, error){
				locationUrl := "https://pokeapi.co/api/v2/location-area/"
				body, error := httpGet(locationUrl+param[0])
				if error{
					return true, nil
				}
				explore := Explore{}
				err := json.Unmarshal(body, &explore)
				if err != nil{
					fmt.Println(err)
					return true, nil
				}
				//unmarshal data view
				intro := fmt.Sprintf("Exploring %s...", param[0])
				fmt.Println(intro)
				for _,poke := range explore.Pokemon_encounters{
					fmt.Println(poke.Pokemon.Name)
				}
				//raw view
				//fmt.Println(string(body))
				return true, nil
			},
		},
	}
}

func httpGet(url string) ([]byte, bool){
	resp, err := http.Get(url)
	errors := false
	if err != nil{
		reply := fmt.Sprintf("unable to go back: %v", err)
		fmt.Println(reply)
		errors = true
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil{
		reply := fmt.Sprintf("Can't read response body: %v", err)
		fmt.Println(reply)
		errors = true
	}
	if resp.StatusCode>299{
		res := fmt.Sprintf("Failure code: %s", resp.StatusCode)
		fmt.Println(res)
		errors = true
	}
	defer resp.Body.Close()
	return body, errors
}

func unmarshal(body []byte, config *listedLocation) bool{
	err := json.Unmarshal(body, config)
	if err != nil{
		reply := fmt.Sprintf("Json Body retrieval error: %v", err)
		fmt.Println(reply)
		return false
	}
	return true
}

func location_print(results []Result){
	for _,location := range results{
		fmt.Println(location.Name)
	}
}



