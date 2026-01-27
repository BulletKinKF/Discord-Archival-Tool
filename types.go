package main

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
