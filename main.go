package main

import (
	"encoding/json"
	"fmt"
	"io"

	// "log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Guild struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Owner       bool     `json:"owner"`
	Permissions string   `json:"permissions"`
	Features    []string `json:"features"`
}

var cfg struct {
	Port int `env:"PORT"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	AuthToken := os.Getenv("authorization")
	println(AuthToken)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(
		"GET",
		"https://discord.com/api/v10/users/@me/guilds",
		nil,
	)

	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", AuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("Discord API error: %d - %s", resp.StatusCode, body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var guilds []Guild
	if err := json.Unmarshal(body, &guilds); err != nil {
		panic(err)
	}

	for _, g := range guilds {
		fmt.Printf("Guild: %s (%s)\n", g.Name, g.ID)
	}
}
