-- https://stackoverflow.com/a/67344658
CREATE TEMPORARY TABLE temp_users AS
SELECT 
    id,
    name,
    password,
    email,
    first_name,
    last_name,
    disabled,
    created_at,
    updated_at
FROM users;

DROP TABLE users;

CREATE TABLE IF NOT EXISTS "users" (
	"id"	INTEGER NOT NULL,
	"name"	TEXT NOT NULL,
	"password"	TEXT NOT NULL,
	"email"	TEXT NOT NULL UNIQUE,
	"first_name"	TEXT,
	"last_name"	TEXT,
	"disabled"	NUMERIC DEFAULT 0,
	"is_oauth"	NUMERIC DEFAULT 0,
	"provider"	TEXT NULL,
	"created_at"	DATETIME,
	"updated_at"	DATETIME,
	PRIMARY KEY("id" AUTOINCREMENT)
);

INSERT INTO users
 (  id,
    name,
    password,
    email,
    first_name,
    last_name,
    disabled,
    created_at,
    updated_at
  )
SELECT
    id,
    name,
    password,
    email,
    first_name,
    last_name,
    disabled,
    created_at,
    updated_at
FROM temp_users;
