package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	var token string

	rootCmd := &cobra.Command{
		Use:   "discord-archiver",
		Short: "A tool for archiving Discord servers",
	}

	// Global --token flag (falls back to env var)
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "Discord authorization token")
	rootCmd.MarkPersistentFlagRequired("token")

	// `archive guilds` — list your guilds
	guildsCmd := &cobra.Command{
		Use:   "guilds",
		Short: "List guilds the token has access to",
		RunE: func(cmd *cobra.Command, args []string) error {
			db := initDatabase()
			arc := NewArchiver(token, db)
			guilds, err := arc.GetGuilds()
			if err != nil {
				return fmt.Errorf("failed to fetch guilds: %w", err)
			}
			for _, g := range guilds {
				fmt.Println(g.ID, " | ", g.Name)
			}
			return nil
		},
	}

	// `archive server <guildID>` — archive a specific server
	archiveCmd := &cobra.Command{
		Use:   "server <guildID>",
		Short: "Archive a specific Discord server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			guildID := args[0]
			db := initDatabase()
			arc := NewArchiver(token, db)
			fmt.Printf("Archiving server: %s\n", guildID)
			// TODO: call your actual archive logic here, e.g. arc.ArchiveGuild(guildID)
			_ = arc
			return nil
		},
	}

	listDiscoveryGuilds := &cobra.Command{
		Use:   "discovery",
		Short: "Lists servers from Discord's discovery feature",
		RunE: func(cmd *cobra.Command, args []string) error {
			db := initDatabase()
			arc := NewArchiver(token, db)
			dg, err := arc.GetDiscoverableGuilds()
			if err != nil {
				return fmt.Errorf("failed to discover guilds: %w", err)
			}
			for _, g := range dg {
				fmt.Printf("Name: %s \nGuild ID: %s \nMember Count: %d \nDescription: %s \n\n", g.Name, g.ID, g.MemberCount, g.Description)
			}
			return nil
		},
	}

	rootCmd.AddCommand(guildsCmd, archiveCmd, listDiscoveryGuilds)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
