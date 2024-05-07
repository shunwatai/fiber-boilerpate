-- Create TRIGGER function for auto update updated_at column after records being altered
CREATE OR REPLACE FUNCTION update_updated_at_column()   
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;   
END;
$$ language 'plpgsql';

-- Create group_users table
DROP SEQUENCE IF EXISTS group_users_id_seq;
CREATE SEQUENCE group_users_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE "public"."group_users" (
    "id" integer DEFAULT nextval('group_users_id_seq') NOT NULL,
    "group_id" integer NOT NULL,
    "user_id"  integer NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "group_users_pkey" PRIMARY KEY ("id")
) WITH (oids = false);

CREATE INDEX "group_users_group_id" ON "group_users" ("group_id");
CREATE INDEX "group_users_user_id" ON "group_users" ("user_id");

ALTER TABLE "group_users"
ADD FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE;
ALTER TABLE "group_users"
ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

-- Apply trigger update_updated_at_column()
CREATE TRIGGER update_group_users_updated_at BEFORE UPDATE ON group_users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
