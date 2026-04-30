package main

type Guild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
	// Features    []string `json:"features"`
}

type Channel struct {
	ID            string `json:"id"`
	Type          int    `json:"type"`
	Name          string `json:"name"`
	Position      int    `json:"position"`
	ParentID      string `json:"parent_id,omitempty"`
	Topic         string `json:"topic,omitempty"`
	GuildID       string `json:"guild_id,omitempty"`
	NSFW          bool   `json:"boolean,omitempty"`
	LastMessageID string `json:"last_message_id,omitempty"`
	Recipients    []User `json:"recipients,omitempty"`
	// TotalMessages int    `json:"total_message_sent,omitempty"` // Even counts the deleted messages. Only availible in threads.
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
	ID            string       `json:"id"`
	Username      string       `json:"username"`
	Discriminator string       `json:"discriminator"`
	Avatar        string       `json:"avatar"`
	Bot           bool         `json:"bot,omitempty"`
	Email         string       `json:"email,omitempty"` // I am guessing this can only be used by Discord employees.
	Flags         int          `json:"flags,omitempty"`
	PremiumTypes  int          `json:"premium_type,omitempty"`
	PrimaryGuild  PrimaryGuild `json:"primary_guild,omitempty"`
}

type GuildMember struct { // Data that is tied to a user within a guild.
	User              User   `json:"user,omitempty"`
	Nickname          string `json:"nick,omitempty"`
	Avatar            string `json:"avatar,omitempty"`        // Hash
	Banner            string `json:"banner,omitempty"`        // Hash
	DateJoined        string `json:"joined_at"`               // ISO8601 timestamp
	PremiumTime       string `json:"premium_since,omitempty"` // ISO8601 timestamp
	Flags             int    `json:"flags"`
	TimeoutExpiration string `json:"communication_disabled_until,omitempty"` // ISO8601 timestamp
}

type PrimaryGuild struct { // This could prove useful for traversing guilds
	GuildId         string `json:"identity_guild_id"`
	IdentityEnabled bool   `json:"identity_enabled"` // null if identity is clear, false if manually removed by guild
	Tag             string `json:"tag"`              // 4 letter text.
	Badge           string `json:"badge"`            // hash
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

type DiscoverableGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
	MemberCount int    `json:"approximate_member_count"`
	// Features    []string `json:"features"`
}
