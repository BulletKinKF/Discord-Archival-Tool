package cmd

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

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Try to load .env if it exists (optional)
	godotenv.Load()

	// Global flags
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "discord_archive.db", "Path to SQLite database")
	rootCmd.PersistentFlags().StringVarP(&authToken, "token", "t", "", "Discord authorization token (required)")
}

// getAuthToken gets the token from flag, env var, or prompts user
func getAuthToken() string {
	token := authToken

	// Priority: 1. Command-line flag, 2. Environment variable, 3. Prompt user
	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
	}

	if token == "" {
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

		fmt.Scanln(&token)
		token = strings.TrimSpace(token)
	}

	if token == "" {
		fmt.Println("❌ Error: authorization token is required")
		fmt.Println()
		fmt.Println("You can provide the token in three ways:")
		fmt.Println("  1. Command line flag:    --token YOUR_TOKEN")
		fmt.Println("  2. Environment variable:  export DISCORD_TOKEN=YOUR_TOKEN")
		fmt.Println("  3. Interactive prompt:    (just run the command without --token)")
		os.Exit(1)
	}

	return token
}
