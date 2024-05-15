CREATE TABLE IF NOT EXISTS  `permission_types` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp,
  `updated_at` datetime NOT NULL DEFAULT current_timestamp ON UPDATE CURRENT_TIMESTAMP
);

-- Pre-populate some default permissions
INSERT INTO `permission_types` (`id`, `name`, `created_at`, `updated_at`) VALUES
(1,	'read',	'2024-05-15 05:45:52',	'2024-05-15 05:45:52'),
(2,	'add',	'2024-05-15 05:45:55',	'2024-05-15 05:45:55'),
(3,	'edit',	'2024-05-15 05:45:58',	'2024-05-15 05:45:58'),
(4,	'delete',	'2024-05-15 05:46:01',	'2024-05-15 05:46:01');
