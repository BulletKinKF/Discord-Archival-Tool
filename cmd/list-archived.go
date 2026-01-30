package cmd

import (
	"fmt"

	"discord-archival-tool/internal/database"

	"github.com/spf13/cobra"
)

var listArchivedCmd = &cobra.Command{
	Use:   "list-archived",
	Short: "List all archived guilds in the database",
	Long:  "Display all guilds that have been archived to the local database.",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.NewDatabase(dbPath)
		if err != nil {
			fmt.Printf("❌ Error initializing database: %v\n", err)
			return
		}
		defer db.Close()

		viewArchivedGuilds(db)
	},
}

func init() {
	rootCmd.AddCommand(listArchivedCmd)
}

func viewArchivedGuilds(db *database.Database) {
	fmt.Println("📚 Archived Guilds:\n")

	guilds, err := db.ListArchivedGuilds()
	if err != nil {
		fmt.Printf("❌ Error fetching archived guilds: %v\n", err)
		return
	}

	if len(guilds) == 0 {
		fmt.Println("ℹ️  No archived guilds found.")
		return
	}

	for i, guild := range guilds {
		fmt.Printf("  %d. %s\n", i+1, guild["name"])
		fmt.Printf("     ID: %s\n", guild["id"])
		fmt.Printf("     Archived: %s\n", guild["archived_at"])
		if i < len(guilds)-1 {
			fmt.Println()
		}
	}
}
