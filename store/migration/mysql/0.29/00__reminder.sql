-- Reminder table for memo reminders
CREATE TABLE `reminder` (
  `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `uid` VARCHAR(32) NOT NULL UNIQUE,
  `memo_id` INT NOT NULL,
  `creator_id` INT NOT NULL,
  `remind_at` BIGINT NOT NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'PENDING',
  `created_ts` BIGINT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
  `updated_ts` BIGINT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
  FOREIGN KEY (`memo_id`) REFERENCES `memo`(`id`) ON DELETE CASCADE,
  FOREIGN KEY (`creator_id`) REFERENCES `user`(`id`) ON DELETE CASCADE
);

CREATE INDEX `idx_reminder_remind_at` ON `reminder`(`remind_at`);
CREATE INDEX `idx_reminder_creator_id` ON `reminder`(`creator_id`);
CREATE INDEX `idx_reminder_status` ON `reminder`(`status`);
CREATE INDEX `idx_reminder_memo_id` ON `reminder`(`memo_id`);
