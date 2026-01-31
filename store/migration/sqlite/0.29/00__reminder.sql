-- Reminder table for memo reminders
CREATE TABLE reminder (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uid TEXT NOT NULL UNIQUE,
  memo_id INTEGER NOT NULL,
  creator_id INTEGER NOT NULL,
  remind_at INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  created_ts INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  updated_ts INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  FOREIGN KEY (memo_id) REFERENCES memo(id) ON DELETE CASCADE,
  FOREIGN KEY (creator_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE INDEX idx_reminder_remind_at ON reminder(remind_at);
CREATE INDEX idx_reminder_creator_id ON reminder(creator_id);
CREATE INDEX idx_reminder_status ON reminder(status);
CREATE INDEX idx_reminder_memo_id ON reminder(memo_id);
