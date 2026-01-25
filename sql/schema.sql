PRAGMA foreign_keys = ON;

/* =========================
   SERVERS (GUILDS)
========================= */

CREATE TABLE IF NOT EXISTS servers (
    id          BIGINT PRIMARY KEY,
    name        TEXT NOT NULL,
    owner_id    BIGINT,
    created_at  TIMESTAMP
);

/* =========================
   CHANNELS
========================= */

CREATE TABLE IF NOT EXISTS channels (
    id          BIGINT PRIMARY KEY,
    server_id   BIGINT NOT NULL,
    name        TEXT,
    type        INTEGER,
    parent_id   BIGINT,
    created_at  TIMESTAMP,

    FOREIGN KEY (server_id) REFERENCES servers(id)
);

/* =========================
   USERS (GLOBAL)
========================= */

CREATE TABLE IF NOT EXISTS users (
    id            BIGINT PRIMARY KEY,
    username      TEXT,
    display_name  TEXT,
    avatar_url    TEXT,
    bot           BOOLEAN
);

/* =========================
   MESSAGES
========================= */

CREATE TABLE IF NOT EXISTS messages (
    id          BIGINT PRIMARY KEY,
    channel_id BIGINT NOT NULL,
    server_id  BIGINT NOT NULL,
    author_id  BIGINT NOT NULL,

    content    TEXT,
    timestamp  TIMESTAMP,
    edited_at  TIMESTAMP,

    pinned     BOOLEAN,
    type       INTEGER,

    raw_json   TEXT,

    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (server_id)  REFERENCES servers(id),
    FOREIGN KEY (author_id)  REFERENCES users(id)
);

/* =========================
   ATTACHMENTS
========================= */

CREATE TABLE IF NOT EXISTS attachments (
    id            BIGINT PRIMARY KEY,
    message_id    BIGINT NOT NULL,

    filename      TEXT,
    url           TEXT,
    proxy_url     TEXT,
    size          INTEGER,
    content_type  TEXT,

    FOREIGN KEY (message_id) REFERENCES messages(id)
);

/* =========================
   EMBEDS
========================= */

CREATE TABLE IF NOT EXISTS embeds (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id BIGINT NOT NULL,

    title       TEXT,
    description TEXT,
    url         TEXT,
    type        TEXT,

    FOREIGN KEY (message_id) REFERENCES messages(id)
);

/* =========================
   REACTIONS
========================= */

CREATE TABLE IF NOT EXISTS reactions (
    message_id BIGINT NOT NULL,
    emoji      TEXT NOT NULL,
    count      INTEGER,

    PRIMARY KEY (message_id, emoji),
    FOREIGN KEY (message_id) REFERENCES messages(id)
);

/* =========================
   MESSAGE EDIT HISTORY
========================= */

CREATE TABLE IF NOT EXISTS message_edits (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id BIGINT NOT NULL,
    old_content TEXT,
    edited_at  TIMESTAMP,

    FOREIGN KEY (message_id) REFERENCES messages(id)
);

/* =========================
   USER MENTIONS
========================= */

CREATE TABLE IF NOT EXISTS mentions (
    message_id BIGINT NOT NULL,
    user_id    BIGINT NOT NULL,

    PRIMARY KEY (message_id, user_id),
    FOREIGN KEY (message_id) REFERENCES messages(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

/* =========================
   INDEXES (PERFORMANCE)
========================= */

CREATE INDEX IF NOT EXISTS idx_channels_server
    ON channels(server_id);

CREATE INDEX IF NOT EXISTS idx_messages_channel_time
    ON messages(channel_id, timestamp);

CREATE INDEX IF NOT EXISTS idx_messages_author
    ON messages(author_id);

CREATE INDEX IF NOT EXISTS idx_messages_server
    ON messages(server_id);

CREATE INDEX IF NOT EXISTS idx_attachments_message
    ON attachments(message_id);

CREATE INDEX IF NOT EXISTS idx_embeds_message
    ON embeds(message_id);

CREATE INDEX IF NOT EXISTS idx_edits_message
    ON message_edits(message_id);
