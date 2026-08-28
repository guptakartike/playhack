ALTER TABLE courts ADD COLUMN IF NOT EXISTS max_players INT DEFAULT 20;

CREATE TABLE IF NOT EXISTS waitlist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'waiting',
    created_at TIMESTAMPTZ DEFAULT now(),
    notified_at TIMESTAMPTZ,
    CONSTRAINT unique_slot_user UNIQUE (slot_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_waitlist_slot_status ON waitlist_entries(slot_id, status);
