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

	fmt.Println("Testing for channels")
	chans, err := arc.GetChannel("1465598299074592906")
	if err != nil {
		panic(err)
	}

	fmt.Println(chans)
}
