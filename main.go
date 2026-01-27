package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Guild struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Owner       bool     `json:"owner"`
	Permissions string   `json:"permissions"`
	Features    []string `json:"features"`
}

type Channel struct {
	ID       string `json:"id"`
	Type     int    `json:"type"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	ParentID string `json:"parent_id,omitempty"`
	Topic    string `json:"topic,omitempty"`
}

type Message struct {
	ID          string       `json:"id"`
	ChannelID   string       `json:"channel_id"`
	Author      User         `json:"author"`
	Content     string       `json:"content"`
	Timestamp   string       `json:"timestamp"`
	Attachments []Attachment `json:"attachments"`
	Embeds      []Embed      `json:"embeds"`
	Mentions    []User       `json:"mentions"`
	Pinned      bool         `json:"pinned"`
	Type        int          `json:"type"`
}

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot,omitempty"`
}

type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int    `json:"size"`
}

type Embed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Color       int    `json:"color,omitempty"`
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) createTables() error {
	schemaBytes, err := os.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	_, err = d.db.Exec(string(schemaBytes))
	return err
}

// Helper function to convert string ID to uint64
func parseID(id string) (uint64, error) {
	return strconv.ParseUint(id, 10, 64)
}

// Helper function to parse optional parent ID
func parseOptionalID(id string) (interface{}, error) {
	if id == "" {
		return nil, nil
	}
	return parseID(id)
}

func (d *Database) SaveGuild(guild *Guild) error {
	guildID, err := parseID(guild.ID)
	if err != nil {
		return fmt.Errorf("invalid guild ID: %w", err)
	}

	featuresJSON, _ := json.Marshal(guild.Features)

	_, err = d.db.Exec(`
		INSERT OR REPLACE INTO guilds (id, name, icon, owner, permissions, features, archived_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, guildID, guild.Name, guild.Icon, guild.Owner, guild.Permissions, string(featuresJSON))

	return err
}

func (d *Database) SaveChannel(channel *Channel, guildID string) error {
	channelID, err := parseID(channel.ID)
	if err != nil {
		return fmt.Errorf("invalid channel ID: %w", err)
	}

	gID, err := parseID(guildID)
	if err != nil {
		return fmt.Errorf("invalid guild ID: %w", err)
	}

	parentID, err := parseOptionalID(channel.ParentID)
	if err != nil {
		return fmt.Errorf("invalid parent ID: %w", err)
	}

	_, err = d.db.Exec(`
		INSERT OR REPLACE INTO channels (id, guild_id, type, name, position, parent_id, topic)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, channelID, gID, channel.Type, channel.Name, channel.Position, parentID, channel.Topic)

	return err
}

func (d *Database) SaveUser(user *User) error {
	userID, err := parseID(user.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	_, err = d.db.Exec(`
		INSERT OR IGNORE INTO users (id, username, discriminator, avatar, bot)
		VALUES (?, ?, ?, ?, ?)
	`, userID, user.Username, user.Discriminator, user.Avatar, user.Bot)

	return err
}

func (d *Database) SaveMessage(message *Message) error {
	messageID, err := parseID(message.ID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	channelID, err := parseID(message.ChannelID)
	if err != nil {
		return fmt.Errorf("invalid channel ID: %w", err)
	}

	authorID, err := parseID(message.Author.ID)
	if err != nil {
		return fmt.Errorf("invalid author ID: %w", err)
	}

	// Save the author first
	if err := d.SaveUser(&message.Author); err != nil {
		return err
	}

	// Save the message
	_, err = d.db.Exec(`
		INSERT OR IGNORE INTO messages (id, channel_id, author_id, content, timestamp, pinned, type)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, messageID, channelID, authorID, message.Content, message.Timestamp, message.Pinned, message.Type)

	if err != nil {
		return err
	}

	// Save attachments
	for _, attachment := range message.Attachments {
		attachmentID, err := parseID(attachment.ID)
		if err != nil {
			return fmt.Errorf("invalid attachment ID: %w", err)
		}

		_, err = d.db.Exec(`
			INSERT OR IGNORE INTO attachments (id, message_id, filename, url, size)
			VALUES (?, ?, ?, ?, ?)
		`, attachmentID, messageID, attachment.Filename, attachment.URL, attachment.Size)

		if err != nil {
			return err
		}
	}

	// Save embeds
	for _, embed := range message.Embeds {
		_, err := d.db.Exec(`
			INSERT INTO embeds (message_id, title, description, url, color)
			VALUES (?, ?, ?, ?, ?)
		`, messageID, embed.Title, embed.Description, embed.URL, embed.Color)

		if err != nil {
			return err
		}
	}

	// Save mentioned users
	for _, mention := range message.Mentions {
		if err := d.SaveUser(&mention); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) GetGuildStats(guildID string) (map[string]interface{}, error) {
	gID, err := parseID(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	stats := make(map[string]interface{})

	// Get guild info
	var guildName string
	var archivedAt string
	err = d.db.QueryRow(`
		SELECT name, archived_at FROM guilds WHERE id = ?
	`, gID).Scan(&guildName, &archivedAt)

	if err != nil {
		return nil, err
	}

	stats["guild_name"] = guildName
	stats["archived_at"] = archivedAt

	// Count channels
	var channelCount int
	d.db.QueryRow(`SELECT COUNT(*) FROM channels WHERE guild_id = ?`, gID).Scan(&channelCount)
	stats["channel_count"] = channelCount

	// Count messages
	var messageCount int
	d.db.QueryRow(`
		SELECT COUNT(*) FROM messages 
		WHERE channel_id IN (SELECT id FROM channels WHERE guild_id = ?)
	`, gID).Scan(&messageCount)
	stats["message_count"] = messageCount

	return stats, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

type Archiver struct {
	client    *http.Client
	authToken string
	baseURL   string
	db        *Database
}

func NewArchiver(authToken string, db *Database) *Archiver {
	return &Archiver{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		authToken: authToken,
		baseURL:   "https://discord.com/api/v10",
		db:        db,
	}
}

func (a *Archiver) makeRequest(method, endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(method, a.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", a.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 429 {
		var rateLimitData struct {
			RetryAfter float64 `json:"retry_after"`
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		json.Unmarshal(body, &rateLimitData)

		waitTime := time.Duration(rateLimitData.RetryAfter*1000) * time.Millisecond
		fmt.Printf("Rate limited. Waiting %.2f seconds...\n", rateLimitData.RetryAfter)
		time.Sleep(waitTime)

		return a.makeRequest(method, endpoint)
	}

	return resp, nil
}

func (a *Archiver) GetGuilds() ([]Guild, error) {
	resp, err := a.makeRequest("GET", "/users/@me/guilds")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var guilds []Guild
	if err := json.Unmarshal(body, &guilds); err != nil {
		return nil, err
	}

	return guilds, nil
}

func (a *Archiver) GetChannels(guildID string) ([]Channel, error) {
	resp, err := a.makeRequest("GET", "/guilds/"+guildID+"/channels")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var channels []Channel
	if err := json.Unmarshal(body, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

func (a *Archiver) GetMessages(channelID string, limit int) ([]Message, error) {
	var allMessages []Message
	var beforeID string

	for {
		endpoint := fmt.Sprintf("/channels/%s/messages?limit=%d", channelID, min(limit-len(allMessages), 100))
		if beforeID != "" {
			endpoint += "&before=" + beforeID
		}

		resp, err := a.makeRequest("GET", endpoint)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
		}

		if err != nil {
			return nil, err
		}

		var messages []Message
		if err := json.Unmarshal(body, &messages); err != nil {
			return nil, err
		}

		if len(messages) == 0 {
			break
		}

		allMessages = append(allMessages, messages...)

		if len(allMessages) >= limit {
			break
		}

		beforeID = messages[len(messages)-1].ID
		time.Sleep(100 * time.Millisecond)
	}

	return allMessages, nil
}

func (a *Archiver) ArchiveGuild(guildID, guildName string) error {
	fmt.Printf("\nArchiving guild: %s\n", guildName)

	// Save guild to database
	guild := &Guild{
		ID:   guildID,
		Name: guildName,
	}
	if err := a.db.SaveGuild(guild); err != nil {
		return err
	}

	// Get and save channels
	channels, err := a.GetChannels(guildID)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		if err := a.db.SaveChannel(&channel, guildID); err != nil {
			return err
		}
	}

	// Archive text channels
	messageCount := 0
	for _, channel := range channels {
		if channel.Type == 0 { // Text channel
			fmt.Printf("  Archiving channel: #%s\n", channel.Name)

			messages, err := a.GetMessages(channel.ID, 10000)
			if err != nil {
				fmt.Printf("    Error: %v\n", err)
				continue
			}

			for _, message := range messages {
				if err := a.db.SaveMessage(&message); err != nil {
					fmt.Printf("    Error saving message: %v\n", err)
					continue
				}
				messageCount++
			}

			fmt.Printf("    Saved %d messages\n", len(messages))
		}
	}

	fmt.Printf("Archive complete: %d channels, %d messages\n", len(channels), messageCount)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	authToken := os.Getenv("authorization")
	if authToken == "" {
		panic("authorization token not set")
	}

	// Initialize database
	db, err := NewDatabase("discord_archive.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	archiver := NewArchiver(authToken, db)

	// Get all guilds
	guilds, err := archiver.GetGuilds()
	if err != nil {
		panic(err)
	}

	fmt.Println("Available guilds:")
	for i, g := range guilds {
		fmt.Printf("%d. %s (%s)\n", i+1, g.Name, g.ID)
	}
	fmt.Println()

	// Archive all guilds
	for _, guild := range guilds {
		if err := archiver.ArchiveGuild(guild.ID, guild.Name); err != nil {
			fmt.Printf("Error archiving %s: %v\n", guild.Name, err)
			continue
		}

		// Show stats
		stats, err := db.GetGuildStats(guild.ID)
		if err == nil {
			fmt.Printf("Stats: %+v\n\n", stats)
		}
	}

	fmt.Println("All archives complete! Database saved to discord_archive.db")
}
