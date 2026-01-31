-- Reminder table for memo reminders
CREATE TABLE reminder (
  id SERIAL PRIMARY KEY,
  uid VARCHAR(32) NOT NULL UNIQUE,
  memo_id INTEGER NOT NULL REFERENCES memo(id) ON DELETE CASCADE,
  creator_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  remind_at BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
  created_ts BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
  updated_ts BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);

CREATE INDEX idx_reminder_remind_at ON reminder(remind_at);
CREATE INDEX idx_reminder_creator_id ON reminder(creator_id);
CREATE INDEX idx_reminder_status ON reminder(status);
CREATE INDEX idx_reminder_memo_id ON reminder(memo_id);
