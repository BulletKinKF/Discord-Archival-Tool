package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func initDatabase() *Database {
	db, err := NewDatabase("discord_archive.db")
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	return db
}

func main() {
	// set up env for testing
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	fmt.Println("Just a quick test")

	db := initDatabase()
	// Execute()

	fmt.Println("Testing.")
	auth := os.Getenv("authorization")
	fmt.Println(auth)
	arc := NewArchiver(auth, db)

	g, err := arc.GetGuilds()
	if err != nil {
		panic(err)
	}

	fmt.Println(g)
}
