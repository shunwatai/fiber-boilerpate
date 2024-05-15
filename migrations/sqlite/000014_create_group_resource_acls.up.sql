CREATE TABLE IF NOT EXISTS "group_resource_acls" (
	"id"	INTEGER NOT NULL,
	"group_id"	NUMERIC,
  "resource_id" NUMERIC,
  "permission_type_id" NUMERIC,
	"created_at"	DATETIME,
	"updated_at"	DATETIME,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("group_id") REFERENCES "groups"("id"),
	FOREIGN KEY("resource_id") REFERENCES "resources"("id"),
	FOREIGN KEY("permission_type_id") REFERENCES "permission_types"("id")
);
