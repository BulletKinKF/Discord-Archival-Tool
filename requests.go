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

func (a *Archiver) GetGuildInfo(guildID string) (*Guild, error) {
	resp, err := a.makeRequest("GET", "/guilds/"+guildID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
	}

	var guild Guild
	if err := json.NewDecoder(resp.Body).Decode(&guild); err != nil {
		return nil, err
	}

	return &guild, nil
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

func (a *Archiver) GetChannelInfo(channelID string) (*Channel, error) {
	resp, err := a.makeRequest("GET", "/channels/"+channelID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
	}

	var channel Channel
	if err := json.NewDecoder(resp.Body).Decode(&channel); err != nil {
		return nil, err
	}

	return &channel, nil
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

// GetMessages fetches all messages from a channel, paginating backwards
// (newest → oldest) using Discord's ?before= parameter.
//
// beforeID seeds the initial cursor — pass the oldest message ID already
// stored to resume an interrupted download, or "" to start from the latest.
// Each page of up to 100 messages is delivered to onBatch immediately so
// callers can persist them before the next page is fetched.
func (a *Archiver) GetMessages(channelID string, beforeID string, onBatch func([]Message) error) (int, error) {
	var totalSaved int

	for {
		endpoint := fmt.Sprintf("/channels/%s/messages?limit=100", channelID)
		if beforeID != "" {
			endpoint += "&before=" + beforeID
		}

		resp, err := a.makeRequest("GET", endpoint)
		if err != nil {
			return totalSaved, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return totalSaved, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
		}

		var messages []Message
		if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
			resp.Body.Close()
			return totalSaved, err
		}
		resp.Body.Close()

		if len(messages) == 0 {
			break
		}

		if err := onBatch(messages); err != nil {
			return totalSaved, err
		}

		totalSaved += len(messages)
		beforeID = messages[len(messages)-1].ID
		time.Sleep(60 * time.Millisecond)
	}

	return totalSaved, nil
}

// SyncNewMessages fetches all messages newer than afterID, paginating
// forward (oldest-of-the-new → newest) using Discord's ?after= parameter.
// Each page is delivered to onBatch immediately, same contract as GetMessages.
func (a *Archiver) SyncNewMessages(channelID string, afterID string, onBatch func([]Message) error) (int, error) {
	var totalSaved int
	for {
		endpoint := fmt.Sprintf("/channels/%s/messages?limit=100", channelID)
		if afterID != "" {
			endpoint += "&after=" + afterID
		}

		resp, err := a.makeRequest("GET", endpoint)
		if err != nil {
			return totalSaved, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return totalSaved, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
		}

		var messages []Message
		if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
			resp.Body.Close()
			return totalSaved, err
		}
		resp.Body.Close()

		if len(messages) == 0 {
			break
		}

		if err := onBatch(messages); err != nil {
			return totalSaved, err
		}
		totalSaved += len(messages)

		// Messages come back newest-first; element [0] is the most recent
		// one in *this* batch, which becomes the floor for the next page.
		afterID = messages[0].ID

		// A short page means we've caught up to the present.
		if len(messages) < 100 {
			break
		}

		time.Sleep(60 * time.Millisecond)
	}
	return totalSaved, nil
}

// ArchiveGuild archives all text channels in a guild.
// For each channel it checks the database for the oldest message already
// stored and resumes downloading from that point, so interrupted runs can
// be safely restarted without re-fetching already-saved messages.
func (a *Archiver) ArchiveGuild(guildID string, progressCallback func(string)) error {
	if progressCallback == nil {
		progressCallback = func(msg string) {}
	}

	guild, err := a.GetGuildInfo(guildID)
	if err != nil {
		return err
	}

	progressCallback(fmt.Sprintf("📦 Archiving guild: %s", guild.Name))

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
		if channel.Type != 0 {
			continue
		}
		textChannelCount++

		// Resume: seed beforeID with the oldest message already saved for
		// this channel. Empty string means no prior data — start fresh.
		resumeID, err := a.db.GetOldestMessageID(channel.ID)
		if err != nil {
			progressCallback(fmt.Sprintf("⚠️ Could not read resume point for #%s: %v", channel.Name, err))
			resumeID = ""
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

		saved, err := a.GetMessages(channel.ID, resumeID, func(batch []Message) error {
			for _, message := range batch {
				if err := a.db.SaveMessage(&message); err != nil {
					progressCallback(fmt.Sprintf("⚠️ Error saving message: %v", err))
				}
			}
			messageCount += len(batch)
			progressCallback(fmt.Sprintf("💾 Saved batch of %d (total %d so far) from #%s", len(batch), messageCount, channel.Name))
			return nil
		})
		if err != nil {
			progressCallback(fmt.Sprintf("⚠️ Error fetching messages from #%s: %v", channel.Name, err))
			continue
		}
		progressCallback(fmt.Sprintf("✅ Finished #%s — %d messages saved this run", channel.Name, saved))

		// New: catch anything posted since the last archive run.
		newestID, err := a.db.GetNewestMessageID(channel.ID)
		if err != nil {
			progressCallback(fmt.Sprintf("⚠️ Could not read newest message for #%s: %v", channel.Name, err))
			newestID = ""
		}
		if newestID != "" {
			newCount, err := a.SyncNewMessages(channel.ID, newestID, func(batch []Message) error {
				for _, message := range batch {
					if err := a.db.SaveMessage(&message); err != nil {
						progressCallback(fmt.Sprintf("⚠️ Error saving message: %v", err))
					}
				}
				return nil
			})
			if err != nil {
				progressCallback(fmt.Sprintf("⚠️ Error syncing new messages in #%s: %v", channel.Name, err))
			} else if newCount > 0 {
				messageCount += newCount
				progressCallback(fmt.Sprintf("🆕 Synced %d new message(s) in #%s", newCount, channel.Name))
			}
		}
	}

	progressCallback(fmt.Sprintf("✨ Archive complete: %d channels, %d messages", len(channels), messageCount))
	return nil
}

func (a *Archiver) ArchiveChannel(channelID string, guildID string, progressCallback func(string)) error {
	if progressCallback == nil {
		progressCallback = func(msg string) {}
	}

	channel, err := a.GetChannelInfo(channelID)
	if err != nil {
		return err
	}

	if channel.GuildID != "" {
		guild, err := a.GetGuildInfo(guildID)
		if err != nil {
			return err
		}

		progressCallback(fmt.Sprintf("💾 Saving parent guild: %s", guild.Name))

		if err := a.db.SaveGuild(guild); err != nil {
			return err
		}
	}

	progressCallback(fmt.Sprintf("📦 Archiving channel: %s", channelID))

	return nil
}

func (a *Archiver) GetDMs() ([]Channel, error) {
	resp, err := a.makeRequest("GET", "/users/@me/channels")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, body)
	}

	fmt.Println(resp.Body)

	var channels []Channel
	if err := json.NewDecoder(resp.Body).Decode(&channels); err != nil {
		return nil, err
	}

	return channels, nil
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
