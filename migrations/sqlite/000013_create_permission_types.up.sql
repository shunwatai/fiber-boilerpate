CREATE TABLE IF NOT EXISTS "permission_types" (
	"id"	INTEGER NOT NULL,
	"name"	TEXT NOT NULL,
	"created_at"	DATETIME NOT NULL,
	"updated_at"	DATETIME NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);

-- Pre-populate some default permissions
INSERT INTO "permission_types" ("id", "name", "created_at", "updated_at") VALUES
(1,	'read',	'2024-05-15 05:41:33.433213+00',	'2024-05-15 05:41:33.433213+00'),
(2,	'add',	'2024-05-15 05:41:36.086894+00',	'2024-05-15 05:41:36.086894+00'),
(3,	'edit',	'2024-05-15 05:41:42.740808+00',	'2024-05-15 05:41:42.740808+00'),
(4,	'delete',	'2024-05-15 05:41:46.279443+00',	'2024-05-15 05:41:46.279443+00');
