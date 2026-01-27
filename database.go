package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

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

func (d *Database) ListArchivedGuilds() ([]map[string]interface{}, error) {
	rows, err := d.db.Query(`
		SELECT id, name, archived_at FROM guilds ORDER BY archived_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guilds []map[string]interface{}
	for rows.Next() {
		var id uint64
		var name, archivedAt string
		if err := rows.Scan(&id, &name, &archivedAt); err != nil {
			return nil, err
		}

		guilds = append(guilds, map[string]interface{}{
			"id":          fmt.Sprintf("%d", id),
			"name":        name,
			"archived_at": archivedAt,
		})
	}

	return guilds, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
