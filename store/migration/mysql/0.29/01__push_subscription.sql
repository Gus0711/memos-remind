-- push_subscription stores web push subscriptions for each user/device
CREATE TABLE `push_subscription` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `user_id` INT NOT NULL,
  `endpoint` TEXT NOT NULL,
  `p256dh` VARCHAR(255) NOT NULL,
  `auth` VARCHAR(255) NOT NULL,
  `user_agent` TEXT NOT NULL,
  `created_ts` INT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_push_subscription_endpoint` (`endpoint`(500)),
  KEY `idx_push_subscription_user_id` (`user_id`),
  CONSTRAINT `fk_push_subscription_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
