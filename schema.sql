-- Discord Archive Database Schema

-- Guilds (Servers)
CREATE TABLE IF NOT EXISTS guilds (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    icon TEXT,
    owner BOOLEAN,
    permissions TEXT,
    archived_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Channels
CREATE TABLE IF NOT EXISTS channels (
    id INTEGER PRIMARY KEY,
    guild_id INTEGER NOT NULL,
    type INTEGER,
    name TEXT NOT NULL,
    position INTEGER,
    parent_id INTEGER,
    topic TEXT,
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    discriminator TEXT,
    avatar TEXT,
    bot BOOLEAN DEFAULT 0
);

-- Messages
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY,
    channel_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    content TEXT,
    pinned BOOLEAN DEFAULT 0,
    type INTEGER,
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (author_id) REFERENCES users(id)
);

-- Attachments
CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY,
    message_id INTEGER NOT NULL,
    filename TEXT,
    url TEXT,
    size INTEGER,
    FOREIGN KEY (message_id) REFERENCES messages(id)
);

-- Embeds
CREATE TABLE IF NOT EXISTS embeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id INTEGER NOT NULL,
    title TEXT,
    description TEXT,
    url TEXT,
    color INTEGER,
    FOREIGN KEY (message_id) REFERENCES messages(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_channels_guild ON channels(guild_id);
CREATE INDEX IF NOT EXISTS idx_messages_channel ON messages(channel_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_attachments_message ON attachments(message_id);