CREATE TABLE IF NOT EXISTS "groups" (
	"id"	INTEGER NOT NULL,
	"name" TEXT UNIQUE,
	"type" TEXT,
	"disabled"	NUMERIC,
	"created_at"	DATETIME,
	"updated_at"	DATETIME,
	PRIMARY KEY("id" AUTOINCREMENT)
);
