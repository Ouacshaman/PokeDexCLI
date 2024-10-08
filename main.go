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
	"math/rand"
)

type cliCommand struct {
	name        string
	description string
	callback    func(map[string]cliCommand, *listedLocation, *pokecache.Cache, []string, *map[string]Pokemon) (bool, error)
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
	Pokemon PokemonEC `json:"pokemon"`
}

type PokemonEC struct{
	Name string `json:"name"`
}

type Pokemon struct{
	Name string `json:"name"`
	Height int `json:"height"`
	Weight int `json:"weight"`
	Stats []Stats `json:"stats"`
	Base_experience int `json:"base_experience"`
	Types []struct{
		Type struct{
			Name string `json:"name"`
		} `json:"type"`
	}`json:"types"`
}

type Stats struct{
	Base_stat int `json:"base_stat"`
	Stat struct {
		Name string `json:"name"`
	} `json:"stat"`
}

func main(){
	active := true
	commands := generate_cmd()
	location := &listedLocation{}
	location.url = "https://pokeapi.co/api/v2/location-area/"
	linkCache := pokecache.NewCache(time.Second*60)
	scanner := bufio.NewScanner(os.Stdin)
	pokedex := make(map[string]Pokemon)
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
		res, err := command.callback(commands, location, linkCache, params, &pokedex)
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
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
			callback:    func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
				return false, nil
			},
		},
		"map": {
			name: "map",
			description: "list locations and explore",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
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
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
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
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
				locationUrl := "https://pokeapi.co/api/v2/location-area/"
				data, found := cache.Get(locationUrl+param[0])
				if found{
					explore := Explore{}
					err := json.Unmarshal(data, &explore)
					if err != nil{
						fmt.Println(err)
						return true, nil
					}
					intro := fmt.Sprintf("Exploring %s...", param[0])
					fmt.Println(intro)
					for _,poke := range explore.Pokemon_encounters{
						fmt.Println(poke.Pokemon.Name)
					}
				} else{
					body, error := httpGet(locationUrl+param[0])
					if error{
						return true, nil
					}
					explore := Explore{}
					cache.Add(locationUrl+param[0], body)
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
				}
				return true, nil
			},
		},
		"catch": {
			name: "catch",
			description: "action created to catch pokemons",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
				baseUrl := "https://pokeapi.co/api/v2/pokemon/"
				completeUrl := baseUrl + param[0]
				body, error := httpGet(completeUrl)
				if error{
					return true, nil
				}
				pokemon := Pokemon{}
				err := json.Unmarshal(body, &pokemon)
				if err != nil{
					return true, nil
				}
				intro := fmt.Sprintf("Throwing a Pokeball at %s...", param[0])
				fmt.Println(intro)
				escaped := fmt.Sprintf("%s escaped!", param[0])
				caught := fmt.Sprintf("%s was caught!", param[0])
				baseXp := pokemon.Base_experience
				rand.Seed(time.Now().UnixNano())
				rand_num := int(rand.Intn(baseXp)/10)
				rand_den := int(baseXp/10)
				rng := rand_num/rand_den
				if rng == 1.0{
					data, ok := (*dex)[param[0]]
					if !ok{
						(*dex)[param[0]] = pokemon
						fmt.Println(caught)
					} else{
						fmt.Println(data.Name,"is already caught")
					}
				} else{
					fmt.Println(escaped)
				}
				return true,nil
			},
		},
		"inspect": {
			name: "inspect",
			description : "Provide stats of Pokemons",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
				data, ok:= (*dex)[param[0]]
				if !ok{
					fmt.Println("you have not caught that pokemon")
					return true, nil
				}
				fmt.Println("Name",data.Name)
				fmt.Println("Height: ",data.Height)
				fmt.Println("Weight: ",data.Weight)
				fmt.Println("Stats:")
				for _,item := range data.Stats{
					output := fmt.Sprintf("  -%s: %d",item.Stat.Name, item.Base_stat)
					fmt.Println(output)
				}
				fmt.Println("Types:")
				fmt.Println("  -",data.Types[0].Type.Name)
				return true, nil
			},
		},
		"pokedex": {
			name: "pokedex",
			description: "list name of all Pokemons caught",
			callback: func(cmds map[string]cliCommand, config *listedLocation, cache *pokecache.Cache, param []string, dex *map[string]Pokemon) (bool, error){
				fmt.Println("Your Pokedex:")
				for i,_ := range (*dex){
					fmt.Println("  -",i)
				}
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



