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
