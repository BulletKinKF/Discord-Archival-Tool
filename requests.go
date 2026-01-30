package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&rateLimitData); err != nil {
			return nil, err
		}

		waitTime := time.Duration(rateLimitData.RetryAfter*1000) * time.Millisecond
		fmt.Printf("⚠️  Rate limited. Waiting %.2f seconds...\n", rateLimitData.RetryAfter)
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

	var guilds []Guild
	if err := json.NewDecoder(resp.Body).Decode(&guilds); err != nil {
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

	var channels []Channel
	if err := json.NewDecoder(resp.Body).Decode(&channels); err != nil {
		return nil, err
	}

	return channels, nil
}

func (a *Archiver) GetUsersFromGuild(guildId string) ([]GuildMember, error) {
	resp, err := a.makeRequest("GET", "/guilds/"+guildId+"/members")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
	}

	var guildmembers []GuildMember
	if err := json.NewDecoder(resp.Body).Decode(&guildmembers); err != nil {
		return nil, err
	}

	return guildmembers, nil
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

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
		}

		var messages []Message
		if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
			return nil, err
		}
		resp.Body.Close()

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

// ArchiveGuild archives a single guild with optional progress callback
func (a *Archiver) ArchiveGuild(guildID, guildName string, progressCallback func(string)) error {
	if progressCallback == nil {
		progressCallback = func(msg string) {} // No-op if not provided
	}

	progressCallback(fmt.Sprintf("📦 Archiving guild: %s", guildName))

	// Save guild to database
	guild := &Guild{
		ID:   guildID,
		Name: guildName,
	}
	if err := a.db.SaveGuild(guild); err != nil {
		return err
	}

	// Get and save channels
	progressCallback("📋 Fetching channels...")
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
	textChannelCount := 0
	for _, channel := range channels {
		if channel.Type == 0 { // Text channel
			textChannelCount++
			progressCallback(fmt.Sprintf("💬 Archiving channel #%s (%d/%d)", channel.Name, textChannelCount, len(channels)))

			messages, err := a.GetMessages(channel.ID, 10000)
			if err != nil {
				progressCallback(fmt.Sprintf("⚠️  Error fetching messages from #%s: %v", channel.Name, err))
				continue
			}

			for _, message := range messages {
				if err := a.db.SaveMessage(&message); err != nil {
					progressCallback(fmt.Sprintf("⚠️  Error saving message: %v", err))
					continue
				}
				messageCount++
			}

			progressCallback(fmt.Sprintf("✅ Saved %d messages from #%s", len(messages), channel.Name))
		}
	}

	progressCallback(fmt.Sprintf("✨ Archive complete: %d channels, %d messages", len(channels), messageCount))
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
