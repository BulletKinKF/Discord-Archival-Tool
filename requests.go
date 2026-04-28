// 50 requests per second to Discord API before Rate limit hits.
// 10,000 requests per 10 minutes results in 24 hour ban on ip. (that's 16.667 requests per sec (fml))
// It is optimal for there to be 15 requests per second. 1 request every 0.06 seconds should suffice.

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

	superProps, err := xSuperProperties()
	if err != nil {
		return nil, fmt.Errorf("failed to generate x-super-properties: %w", err)
	}

	req.Header.Set("Authorization", a.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-super-properties", superProps)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 429 {
		var rateLimitData struct {
			Message    string  `json:"message"`
			RetryAfter float64 `json:"retry_after"`
			Global     bool    `json:"boolean"`
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&rateLimitData); err != nil {
			return nil, err
		}

		waitTime := time.Duration(rateLimitData.RetryAfter*1000) * time.Millisecond
		fmt.Printf("⚠️ Rate limited. Waiting %.2f seconds...\n", rateLimitData.RetryAfter)
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

	// Bug fix: body was being read twice — once with ReadAll (and printed),
	// then again with json.NewDecoder, which would always get an empty reader.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))

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

	var channels []Channel
	if err := json.NewDecoder(resp.Body).Decode(&channels); err != nil {
		return nil, err
	}

	return channels, nil
}

// I don't think this function works...
func (a *Archiver) GetUsersFromGuild(guildId string) ([]GuildMember, error) {
	resp, err := a.makeRequest("GET", "/guilds/"+guildId+"/members?limit=1000")
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

// GetMessages fetches up to `limit` messages from a channel, paginating
// backwards (newest → oldest) using Discord's `?before=` parameter.
//
// resumeBeforeID, if non-empty, is used as the initial `before` cursor.
// This lets callers resume an interrupted download by passing in the oldest
// message ID they already have stored — fetching will continue from that
// point backwards into older history without re-downloading anything already
// saved. Pass an empty string to start from the very latest message.
func (a *Archiver) GetMessages(channelID string, limit int, resumeBeforeID string) ([]Message, error) {
	var allMessages []Message
	beforeID := resumeBeforeID // start at the resume point (may be "")

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
			resp.Body.Close()
			return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
		}

		var messages []Message
		if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
			resp.Body.Close()
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
		time.Sleep(60 * time.Millisecond) // 0.06 seconds
	}

	return allMessages, nil
}

// ArchiveGuild archives all text channels in a guild.
// For each channel it first checks the database for the oldest message
// already stored and resumes downloading from that point, so interrupted
// runs can be safely restarted without re-fetching messages that are already
// saved.
func (a *Archiver) ArchiveGuild(guildID, guildName string, progressCallback func(string)) error {
	if progressCallback == nil {
		progressCallback = func(msg string) {}
	}

	progressCallback(fmt.Sprintf("📦 Archiving guild: %s", guildName))

	guild := &Guild{
		ID:   guildID,
		Name: guildName,
	}
	if err := a.db.SaveGuild(guild); err != nil {
		return err
	}

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

	messageCount := 0
	textChannelCount := 0

	for _, channel := range channels {
		textChannelCount++

		// Check whether we have already started archiving this channel.
		// GetOldestMessageID returns "" when the channel has no stored
		// messages, which causes GetMessages to start from the latest
		// message — the normal first-run behaviour.
		resumeID, err := a.db.GetOldestMessageID(channel.ID)
		if err != nil {
			progressCallback(fmt.Sprintf("⚠️ Could not read resume point for #%s: %v", channel.Name, err))
			resumeID = "" // fall back to a fresh fetch
		}

		if resumeID != "" {
			progressCallback(fmt.Sprintf(
				"⏩ Resuming #%s from message ID %s (%d/%d)",
				channel.Name, resumeID, textChannelCount, len(channels),
			))
		} else {
			progressCallback(fmt.Sprintf(
				"💬 Archiving channel #%s (%d/%d)",
				channel.Name, textChannelCount, len(channels),
			))
		}

		messages, err := a.GetMessages(channel.ID, 10000, resumeID)
		if err != nil {
			progressCallback(fmt.Sprintf("⚠️ Error fetching messages from #%s: %v", channel.Name, err))
			continue
		}

		for _, message := range messages {
			if err := a.db.SaveMessage(&message); err != nil {
				progressCallback(fmt.Sprintf("⚠️ Error saving message: %v", err))
				continue
			}
			messageCount++
		}

		progressCallback(fmt.Sprintf("✅ Saved %d messages from #%s", len(messages), channel.Name))
	}

	progressCallback(fmt.Sprintf("✨ Archive complete: %d channels, %d messages", len(channels), messageCount))
	return nil
}

// Misc Functions

func (a *Archiver) GetDiscoverableGuilds() ([]DiscoverableGuild, error) {
	resp, err := a.makeRequest("GET", "/discoverable-guilds")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
	}

	var result struct {
		DiscoveredGuilds []DiscoverableGuild `json:"guilds"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.DiscoveredGuilds, nil
}
