CREATE TABLE IF NOT EXISTS "resources" (
	"id"	INTEGER NOT NULL,
	"name"	TEXT,
	"disabled"	NUMERIC DEFAULT 0,
	"created_at"	DATETIME,
	"updated_at"	DATETIME,
	PRIMARY KEY("id" AUTOINCREMENT)
);

-- Pre-populate some default resources
INSERT INTO "resources" ("id", "name", "disabled", "created_at", "updated_at") VALUES
(1,	'users',	'0',	'2024-05-14 06:54:25.780889+00',	'2024-05-14 06:54:25.780889+00'),
(2,	'groups',	'0',	'2024-05-14 06:54:30.014063+00',	'2024-05-14 06:54:30.014063+00'),
(3,	'todos',	'0',	'2024-05-14 06:54:33.907284+00',	'2024-05-14 06:54:33.907284+00');
