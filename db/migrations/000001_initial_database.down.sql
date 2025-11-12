-- ================================
-- INDEXES
-- ================================
DROP INDEX IF EXISTS idx_messages_room_id_sent_at_unblocked;
DROP INDEX IF EXISTS idx_messages_sent_time;
DROP INDEX IF EXISTS idx_messages_sender_id;
DROP INDEX IF EXISTS idx_messages_room_id_sent_at;

DROP INDEX IF EXISTS idx_room_participants_user_id;

DROP INDEX IF EXISTS idx_rooms_livestream_id;
DROP INDEX IF EXISTS idx_rooms_type;
DROP INDEX IF EXISTS idx_rooms_owner_id;

-- ================================
-- TRIGGERS & FUNCTIONS
-- ================================
DROP TRIGGER IF EXISTS trig_update_last_message ON messages;
DROP FUNCTION IF EXISTS update_last_message_id;

DROP TRIGGER IF EXISTS trig_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at;

-- ================================
-- TABLES
-- ================================
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS room_participants;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;

-- ================================
-- TYPES
-- ================================
DROP TYPE IF EXISTS room_type;
