-- Create TRIGGER function for auto update updated_at column after records being altered
CREATE OR REPLACE FUNCTION update_updated_at_column()   
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;   
END;
$$ language 'plpgsql';

-- Create group_resource_acls table
DROP SEQUENCE IF EXISTS group_resource_acls_id_seq;
CREATE SEQUENCE group_resource_acls_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE "public"."group_resource_acls" (
    "id" integer DEFAULT nextval('group_resource_acls_id_seq') NOT NULL,
    "group_id" integer NOT NULL,
    "resource_id" integer NOT NULL,
    "permission_type_id" integer NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "group_resource_acls_pkey" PRIMARY KEY ("id")
) WITH (oids = false);

CREATE INDEX "group_resource_acls_group_id" ON "group_resource_acls" ("group_id");
CREATE INDEX "group_resource_acls_resource_id" ON "group_resource_acls" ("resource_id");
CREATE INDEX "group_resource_acls_permission_type_id" ON "group_resource_acls" ("permission_type_id");

ALTER TABLE "group_resource_acls"
ADD FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "group_resource_acls"
ADD FOREIGN KEY ("resource_id") REFERENCES "resources" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "group_resource_acls"
ADD FOREIGN KEY ("permission_type_id") REFERENCES "permission_types" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- Apply trigger update_updated_at_column()
CREATE TRIGGER update_group_resource_acls_updated_at BEFORE UPDATE ON group_resource_acls FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
