package cmd

import (
	"fmt"
	"strings"

	"discord-archival-tool/internal/database"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [guild-id]",
	Short: "View statistics for archived guilds",
	Long:  "Display statistics such as message count, channel count, and archive date.",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.NewDatabase(dbPath)
		if err != nil {
			fmt.Printf("❌ Error initializing database: %v\n", err)
			return
		}
		defer db.Close()

		if len(args) > 0 {
			viewGuildStatsByID(db, args[0])
		} else {
			viewGuildStatsInteractive(db)
		}
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func viewGuildStatsByID(db *database.Database, guildID string) {
	stats, err := db.GetGuildStats(guildID)
	if err != nil {
		fmt.Printf("❌ Error fetching stats: %v\n", err)
		return
	}

	fmt.Printf("\n" + strings.Repeat("═", 40) + "\n")
	fmt.Printf("📊 Statistics for '%s'\n", stats["guild_name"])
	fmt.Printf(strings.Repeat("─", 40) + "\n")
	fmt.Printf("  Channels:  %d\n", stats["channel_count"])
	fmt.Printf("  Messages:  %d\n", stats["message_count"])
	fmt.Printf("  Archived:  %s\n", stats["archived_at"])
	fmt.Printf(strings.Repeat("═", 40) + "\n")
}

func viewGuildStatsInteractive(db *database.Database) {
	fmt.Println("📊 View Guild Statistics\n")

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
	}

	fmt.Print("\n➤ Enter guild number (or 0 to cancel): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice < 1 || choice > len(guilds) {
		fmt.Println("❌ Invalid choice.")
		return
	}

	selectedGuild := guilds[choice-1]
	guildID := selectedGuild["id"].(string)

	viewGuildStatsByID(db, guildID)
}
