package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  Warning: .env file not found")
	}

	authToken := os.Getenv("authorization")
	if authToken == "" {
		fmt.Println("❌ Error: authorization token not set in .env file")
		os.Exit(1)
	}

	// Initialize database
	db, err := NewDatabase("discord_archive.db")
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize archiver
	archiver := NewArchiver(authToken, db)

	// Start CLI
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║   Discord Server Archiver - CLI v1.0  ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	for {
		printMenu()
		fmt.Print("\n➤ Enter choice: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			listGuilds(archiver)
		case "2":
			archiveGuild(archiver, scanner)
		case "3":
			archiveAllGuilds(archiver)
		case "4":
			viewArchivedGuilds(db)
		case "5":
			viewGuildStats(db, scanner)
		case "6":
			fmt.Println("\n👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid choice. Please try again.")
		}
	}
}

func printMenu() {
	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Println("📋 Menu:")
	fmt.Println("  1. List available guilds")
	fmt.Println("  2. Archive a specific guild")
	fmt.Println("  3. Archive all guilds")
	fmt.Println("  4. View archived guilds")
	fmt.Println("  5. View guild statistics")
	fmt.Println("  6. Exit")
	fmt.Println(strings.Repeat("─", 40))
}

func listGuilds(archiver *Archiver) {
	fmt.Println("\n📡 Fetching guilds from Discord...")

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
		ownerBadge := ""
		if guild.Owner {
			ownerBadge = " 👑"
		}
		fmt.Printf("  %d. %s%s\n", i+1, guild.Name, ownerBadge)
		fmt.Printf("     ID: %s\n", guild.ID)
	}
}

func archiveGuild(archiver *Archiver, scanner *bufio.Scanner) {
	fmt.Println("\n📡 Fetching guilds from Discord...")

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
	if !scanner.Scan() {
		return
	}

	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 0 || choice > len(guilds) {
		fmt.Println("❌ Invalid choice.")
		return
	}

	if choice == 0 {
		fmt.Println("❌ Cancelled.")
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
	fmt.Println("\n📡 Fetching guilds from Discord...")

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

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("❌ Cancelled.")
		return
	}

	fmt.Println("\n🚀 Starting archive process...\n")

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

	fmt.Printf("\n" + strings.Repeat("═", 40) + "\n")
	fmt.Printf("✅ Archive complete!\n")
	fmt.Printf("   Success: %d\n", successCount)
	if failCount > 0 {
		fmt.Printf("   Failed: %d\n", failCount)
	}
	fmt.Printf(strings.Repeat("═", 40) + "\n")
}

func viewArchivedGuilds(db *Database) {
	fmt.Println("\n📚 Archived Guilds:\n")

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

func viewGuildStats(db *Database, scanner *bufio.Scanner) {
	fmt.Println("\n📊 View Guild Statistics\n")

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
	if !scanner.Scan() {
		return
	}

	choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || choice < 0 || choice > len(guilds) {
		fmt.Println("❌ Invalid choice.")
		return
	}

	if choice == 0 {
		fmt.Println("❌ Cancelled.")
		return
	}

	selectedGuild := guilds[choice-1]
	guildID := selectedGuild["id"].(string)

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
