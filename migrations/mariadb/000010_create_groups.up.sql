CREATE TABLE IF NOT EXISTS  `groups` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(255) UNIQUE NOT NULL,
  `type` varchar(255) NOT NULL,
  `disabled` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT current_timestamp,
  `updated_at` datetime NOT NULL DEFAULT current_timestamp ON UPDATE CURRENT_TIMESTAMP
);

-- Pre-populate default admin group
INSERT INTO `groups` (`id`, `name`, `type`, `disabled`, `created_at`, `updated_at`) VALUES
(1,	'admin', 'admin', '0',	'2024-05-14 06:54:25',	'2024-05-14 06:54:25');
