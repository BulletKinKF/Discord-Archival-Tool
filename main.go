package main

import (
	"fmt"
	"os"
)

func initDatabase() *Database {
	db, err := NewDatabase(dbPath)
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	return db
}

func main() {

	// Execute()
}
