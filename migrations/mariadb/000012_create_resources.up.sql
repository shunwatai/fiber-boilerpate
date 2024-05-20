CREATE TABLE IF NOT EXISTS  `resources` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(255) NOT NULL,
  `order` int NOT NULL,
  `disabled` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT current_timestamp,
  `updated_at` datetime NOT NULL DEFAULT current_timestamp ON UPDATE CURRENT_TIMESTAMP
);

-- Pre-populate some default resources
INSERT INTO `resources` (`id`, `name`, `order`, `disabled`, `created_at`, `updated_at`) VALUES
(1,	'users',	1, 0,	'2024-05-14 06:57:55',	'2024-05-14 06:57:55'),
(2,	'groups',	2, 0,	'2024-05-14 06:57:59',	'2024-05-14 06:57:59'),
(3,	'todos',	3, 0,	'2024-05-14 06:58:03',	'2024-05-14 06:58:03');
