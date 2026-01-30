package cmd

import (
	"fmt"
	"strings"

	"discord-archival-tool/internal/archiver"
	"discord-archival-tool/internal/database"

	"github.com/spf13/cobra"
)

var (
	archiveAll   bool
	messageLimit int
)

var archiveCmd = &cobra.Command{
	Use:   "archive [guild-id]",
	Short: "Archive a specific guild or all guilds",
	Long:  "Archive Discord server data including channels, messages, and users.",
	Run: func(cmd *cobra.Command, args []string) {
		token := getAuthToken()
		db, err := database.NewDatabase(dbPath)
		if err != nil {
			fmt.Printf("❌ Error initializing database: %v\n", err)
			return
		}
		defer db.Close()

		arch := archiver.NewArchiver(token, db)

		if archiveAll {
			archiveAllGuilds(arch)
		} else if len(args) > 0 {
			archiveGuildByID(arch, args[0])
		} else {
			archiveInteractive(arch)
		}
	},
}

func init() {
	archiveCmd.Flags().BoolVarP(&archiveAll, "all", "a", false, "Archive all available guilds")
	archiveCmd.Flags().IntVarP(&messageLimit, "limit", "l", 10000, "Maximum number of messages per channel")
	rootCmd.AddCommand(archiveCmd)
}

func archiveGuildByID(arch *archiver.Archiver, guildID string) {
	guilds, err := arch.GetGuilds()
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

	err = arch.ArchiveGuild(guildID, guildName, func(msg string) {
		fmt.Println("  " + msg)
	})

	if err != nil {
		fmt.Printf("\n❌ Error archiving guild: %v\n", err)
		return
	}

	fmt.Println("\n✅ Archive completed successfully!")
}

func archiveInteractive(arch *archiver.Archiver) {
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

	err = arch.ArchiveGuild(selectedGuild.ID, selectedGuild.Name, func(msg string) {
		fmt.Println("  " + msg)
	})

	if err != nil {
		fmt.Printf("\n❌ Error archiving guild: %v\n", err)
		return
	}

	fmt.Println("\n✅ Archive completed successfully!")
}

func archiveAllGuilds(arch *archiver.Archiver) {
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

	fmt.Printf("\n⚠️  About to archive %d guild(s). This may take a while.\n", len(guilds))
	fmt.Print("➤ Continue? (y/n): ")

	var confirm string
	fmt.Scanf("%s", &confirm)
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("❌ Cancelled.")
		return
	}

	fmt.Println("\n🚀 Starting archive process...\n")

	successCount := 0
	failCount := 0

	for i, guild := range guilds {
		fmt.Printf("\n[%d/%d] Archiving '%s'...\n", i+1, len(guilds), guild.Name)

		err := arch.ArchiveGuild(guild.ID, guild.Name, func(msg string) {
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
