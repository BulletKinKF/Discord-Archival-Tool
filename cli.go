package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	dbPath    string
	authToken string
)

var rootCmd = &cobra.Command{
	Use:   "discord-archiver",
	Short: "Archive Discord servers to SQLite database",
	Long: `Discord Server Archiver - A tool to backup Discord servers including 
channels, messages, users, and attachments to a local SQLite database.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available guilds from Discord",
	Long:  "Fetches and displays all guilds (servers) you have access to on Discord.",
	Run: func(cmd *cobra.Command, args []string) {
		archiver := initArchiver()
		listGuilds(archiver)
	},
}

var archiveCmd = &cobra.Command{
	Use:   "archive [guild-id]",
	Short: "Archive a specific guild or all guilds",
	Long:  "Archive Discord server data including channels, messages, and users.",
	Run: func(cmd *cobra.Command, args []string) {
		archiver := initArchiver()
		all, _ := cmd.Flags().GetBool("all")

		if all {
			archiveAllGuilds(archiver)
		} else if len(args) > 0 {
			archiveGuildByID(archiver, args[0])
		} else {
			archiveInteractive(archiver)
		}
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats [guild-id]",
	Short: "View statistics for archived guilds",
	Long:  "Display statistics such as message count, channel count, and archive date.",
	Run: func(cmd *cobra.Command, args []string) {
		db := initDatabase()
		defer db.Close()

		if len(args) > 0 {
			viewGuildStatsByID(db, args[0])
		} else {
			viewArchivedGuilds(db)
		}
	},
}

var listArchivedCmd = &cobra.Command{
	Use:   "list-archived",
	Short: "List all archived guilds in the database",
	Long:  "Display all guilds that have been archived to the local database.",
	Run: func(cmd *cobra.Command, args []string) {
		db := initDatabase()
		defer db.Close()
		viewArchivedGuilds(db)
	},
}

func init() {
	// Try to load .env if it exists (optional)
	godotenv.Load()

	// Global flags
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "discord_archive.db", "Path to SQLite database")
	rootCmd.PersistentFlags().StringVarP(&authToken, "token", "t", "", "Discord authorization token (required)")

	// Archive command flags
	archiveCmd.Flags().BoolP("all", "a", false, "Archive all available guilds")
	archiveCmd.Flags().IntP("limit", "l", 10000, "Maximum number of messages per channel")

	// Add commands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(archiveCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(listArchivedCmd)
}

func initDatabase() *Database {
	db, err := NewDatabase(dbPath)
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	return db
}

func initArchiver() *Archiver {
	// Priority: 1. Command-line flag, 2. Environment variable, 3. Prompt user
	if authToken == "" {
		authToken = os.Getenv("DISCORD_TOKEN")
	}

	if authToken == "" {
		fmt.Println("╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║  Discord Authorization Token Required                         ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Println("📝 How to get your Discord token:")
		fmt.Println("   1. Open Discord in your web browser")
		fmt.Println("   2. Press F12 to open Developer Tools")
		fmt.Println("   3. Go to the 'Network' tab")
		fmt.Println("   4. Press F5 to reload the page")
		fmt.Println("   5. Click any request and look for 'authorization' in Headers")
		fmt.Println("   6. Copy the token value")
		fmt.Println()
		fmt.Println("⚠️  Keep your token private! Never share it with anyone.")
		fmt.Println()
		fmt.Print("➤ Enter your Discord token: ")

		fmt.Scanln(&authToken)
		authToken = strings.TrimSpace(authToken)
	}

	if authToken == "" {
		fmt.Println("❌ Error: authorization token is required")
		fmt.Println()
		fmt.Println("You can provide the token in three ways:")
		fmt.Println("  1. Command line flag:    --token YOUR_TOKEN")
		fmt.Println("  2. Environment variable:  export DISCORD_TOKEN=YOUR_TOKEN")
		fmt.Println("  3. Interactive prompt:    (just run the command without --token)")
		os.Exit(1)
	}

	db := initDatabase()
	return NewArchiver(authToken, db)
}

func listGuilds(archiver *Archiver) {
	fmt.Println("📡 Fetching guilds from Discord...")

	guilds, err := archiver.GetGuilds()
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
}

func archiveGuildByID(archiver *Archiver, guildID string) {
	// First, fetch guild info to get the name
	guilds, err := archiver.GetGuilds()
	if err != nil {
		fmt.Printf("❌ Error fetching guilds: %v\n", err)
		return
	}

	var guildName string
	found := false
	for _, guild := range guilds {
		if guild.ID == guildID {
			guildName = guild.Name
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ Guild with ID '%s' not found\n", guildID)
		return
	}

	fmt.Printf("🚀 Starting archive of '%s'...\n\n", guildName)

	err = archiver.ArchiveGuild(guildID, guildName, func(msg string) {
		fmt.Println("  " + msg)
	})

	if err != nil {
		fmt.Printf("\n❌ Error archiving guild: %v\n", err)
		return
	}

	fmt.Println("\n✅ Archive completed successfully!")
}

func archiveInteractive(archiver *Archiver) {
	fmt.Println("📡 Fetching guilds from Discord...")

	guilds, err := archiver.GetGuilds()
	if err != nil {
		fmt.Printf("❌ Error fetching guilds: %v\n", err)
		return
	}

	if len(guilds) == 0 {
		fmt.Println("ℹ️  No guilds found.")
		return
	}

	fmt.Printf("\n✅ Found %d guild(s):\n\n", len(guilds))
	for i, guild := range guilds {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, guild.Name, guild.ID)
	}

	fmt.Print("\n➤ Enter guild number to archive (or 0 to cancel): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice < 1 || choice > len(guilds) {
		fmt.Println("❌ Invalid choice.")
		return
	}

	selectedGuild := guilds[choice-1]

	fmt.Printf("\n🚀 Starting archive of '%s'...\n\n", selectedGuild.Name)

	err = archiver.ArchiveGuild(selectedGuild.ID, selectedGuild.Name, func(msg string) {
		fmt.Println("  " + msg)
	})

	if err != nil {
		fmt.Printf("\n❌ Error archiving guild: %v\n", err)
		return
	}

	fmt.Println("\n✅ Archive completed successfully!")
}

func archiveAllGuilds(archiver *Archiver) {
	fmt.Println("📡 Fetching guilds from Discord...")

	guilds, err := archiver.GetGuilds()
	if err != nil {
		fmt.Printf("❌ Error fetching guilds: %v\n", err)
		return
	}

	if len(guilds) == 0 {
		fmt.Println("ℹ️  No guilds found.")
		return
	}

	fmt.Printf("\n⚠️  About to archive %d guild(s). This may take a while.\n", len(guilds))
	fmt.Print("➤ Continue? (y/n): ")

	var confirm string
	fmt.Scanf("%s", &confirm)
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("❌ Cancelled.")
		return
	}

	fmt.Println("\n🚀 Starting archive process...")

	successCount := 0
	failCount := 0

	for i, guild := range guilds {
		fmt.Printf("\n[%d/%d] Archiving '%s'...\n", i+1, len(guilds), guild.Name)

		err := archiver.ArchiveGuild(guild.ID, guild.Name, func(msg string) {
			fmt.Println("  " + msg)
		})

		if err != nil {
			fmt.Printf("❌ Failed to archive '%s': %v\n", guild.Name, err)
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("\n" + "════════════════════════════════════════" + "\n")
	fmt.Printf("✅ Archive complete!\n")
	fmt.Printf("   Success: %d\n", successCount)
	if failCount > 0 {
		fmt.Printf("   Failed: %d\n", failCount)
	}
	fmt.Printf("\n" + "════════════════════════════════════════" + "\n")
}

func viewArchivedGuilds(db *Database) {
	fmt.Println("📚 Archived Guilds:")

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

func viewGuildStatsByID(db *Database, guildID string) {
	stats, err := db.GetGuildStats(guildID)
	if err != nil {
		fmt.Printf("❌ Error fetching stats: %v\n", err)
		return
	}

	fmt.Printf("\n" + "════════════════════════════════════════" + "\n")
	fmt.Printf("📊 Statistics for '%s'\n", stats["guild_name"])
	fmt.Printf("────────────────────────────────────────" + "\n")
	fmt.Printf("  Channels:  %d\n", stats["channel_count"])
	fmt.Printf("  Messages:  %d\n", stats["message_count"])
	fmt.Printf("  Archived:  %s\n", stats["archived_at"])
	fmt.Printf("════════════════════════════════════════" + "\n")
}

func viewGuildStatsInteractive(db *Database) {
	fmt.Println("📊 View Guild Statistics")

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

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
