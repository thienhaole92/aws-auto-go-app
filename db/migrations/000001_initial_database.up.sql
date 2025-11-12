-- ================================
-- Extensions
-- ================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ================================
-- USERS
-- ================================
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    avatar_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Trigger function to auto-update updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- ================================
-- ROOMS
-- ================================
CREATE TYPE room_type AS ENUM ('private', 'livestream');

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type room_type NOT NULL,
    owner_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_message_id UUID,
    livestream_id TEXT NULL
);

ALTER TABLE rooms
ADD CONSTRAINT uq_rooms_livestream_id UNIQUE (livestream_id);

-- ================================
-- ROOM PARTICIPANTS
-- ================================
CREATE TABLE room_participants (
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
);

-- ================================
-- MESSAGES
-- ================================
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    sent_at BIGINT NOT NULL DEFAULT (extract(epoch from now())::BIGINT),
    read_at TIMESTAMPTZ NULL,
    is_emoji BOOLEAN DEFAULT FALSE,
    category INTEGER,
    is_blocked BOOLEAN DEFAULT FALSE,
    blocked_reason TEXT
);

-- ================================
-- TRIGGER: Update last_message_id on new messages
-- ================================
CREATE OR REPLACE FUNCTION update_last_message_id()
RETURNS TRIGGER AS $$
BEGIN
    -- Only update if this is the newest message
    UPDATE rooms
    SET last_message_id = NEW.id
    WHERE id = NEW.room_id
        AND (
            last_message_id IS NULL OR
            NEW.sent_at >= (
                SELECT sent_at FROM messages WHERE id = last_message_id
            )
        );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_update_last_message
AFTER INSERT ON messages
FOR EACH ROW EXECUTE PROCEDURE update_last_message_id();

-- ================================
-- INDEXES
-- ================================
-- Rooms
CREATE INDEX idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX idx_rooms_type ON rooms(type);
CREATE INDEX idx_rooms_livestream_id ON rooms(livestream_id);

-- Room participants
CREATE INDEX idx_room_participants_user_id ON room_participants(user_id);

-- Messages
CREATE INDEX idx_messages_room_id_sent_at ON messages(room_id, sent_at DESC);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_sent_time ON messages(sent_at DESC);

-- Partial index for unblocked messages
CREATE INDEX idx_messages_room_id_sent_at_unblocked
ON messages(room_id, sent_at DESC)
WHERE is_blocked = FALSE;
