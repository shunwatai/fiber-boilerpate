CREATE TABLE IF NOT EXISTS  `group_users` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `group_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp,
  `updated_at` datetime NOT NULL DEFAULT current_timestamp ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
);

-- Pre-populate admin user assigns into admin group
INSERT INTO `group_users` (`id`, `group_id`, `user_id`, `created_at`, `updated_at`) VALUES
(1,	1, 1,	'2024-05-14 06:54:25',	'2024-05-14 06:54:25');
