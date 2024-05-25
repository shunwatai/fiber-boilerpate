CREATE TABLE IF NOT EXISTS "group_users" (
	"id"	INTEGER NOT NULL,
	"group_id"	NUMERIC,
  "user_id"  	NUMERIC,
	"created_at"	DATETIME,
	"updated_at"	DATETIME,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("group_id") REFERENCES "groups"("id"),
	FOREIGN KEY("user_id") REFERENCES "users"("id")
);

-- Pre-populate admin user assigns into admin group
INSERT INTO "group_users" ("id", "group_id", "user_id", "created_at", "updated_at") VALUES
(1,	1, 1,	'2024-05-14 06:54:25.780889+00',	'2024-05-14 06:54:25.780889+00');
