package cmd

import (
	"fmt"

	"discord-archival-tool/internal/archiver"
	"discord-archival-tool/internal/database"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available guilds from Discord",
	Long:  "Fetches and displays all guilds (servers) you have access to on Discord.",
	Run: func(cmd *cobra.Command, args []string) {
		token := getAuthToken()
		db, err := database.NewDatabase(dbPath)
		if err != nil {
			fmt.Printf("❌ Error initializing database: %v\n", err)
			return
		}
		defer db.Close()

		arch := archiver.NewArchiver(token, db)

		fmt.Println("📡 Fetching guilds from Discord...")

		guilds, err := arch.GetGuilds()
		if err != nil {
			fmt.Printf("❌ Error fetching guilds: %v\n", err)
			return
		}

		if len(guilds) == 0 {
			fmt.Println("ℹ️  No guilds found.")
			return
		}

		fmt.Printf("\n✅ Found %d guild(s):\n\n", len(guilds))
		for _, guild := range guilds {
			ownerBadge := ""
			if guild.Owner {
				ownerBadge = " 👑"
			}
			fmt.Printf("  • %s%s\n", guild.Name, ownerBadge)
			fmt.Printf("    ID: %s\n", guild.ID)
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
