CREATE TABLE IF NOT EXISTS "groups" (
	"id"	INTEGER NOT NULL,
	"name" TEXT UNIQUE,
	"type" TEXT,
	"disabled"	NUMERIC,
	"created_at"	DATETIME,
	"updated_at"	DATETIME,
	PRIMARY KEY("id" AUTOINCREMENT)
);

-- Pre-populate default admin group
INSERT INTO "groups" ("id", "name", "type", "disabled", "created_at", "updated_at") VALUES
(1,	'admin', 'admin', '0',	'2024-05-14 06:54:25.780889+00',	'2024-05-14 06:54:25.780889+00');
