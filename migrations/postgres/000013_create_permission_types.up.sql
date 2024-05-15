-- Create TRIGGER function for auto update updated_at column after records being altered
CREATE OR REPLACE FUNCTION update_updated_at_column()   
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;   
END;
$$ language 'plpgsql';

-- Create permission_types table
DROP SEQUENCE IF EXISTS permission_types_id_seq;
CREATE SEQUENCE permission_types_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE "public"."permission_types" (
    "id" integer DEFAULT nextval('permission_types_id_seq') NOT NULL,
    "name" character varying(255) NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "permission_types_pkey" PRIMARY KEY ("id")
) WITH (oids = false);
COMMENT ON COLUMN "permission_types"."name" IS 'permission attribute like add, read, edit, delete';

-- Apply trigger update_updated_at_column()
CREATE TRIGGER update_permission_types_updated_at BEFORE UPDATE ON permission_types FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Pre-populate some default permissions
INSERT INTO "permission_types" ("id", "name", "created_at", "updated_at") VALUES
(1,	'read',	'2024-05-15 05:41:33.433213+00',	'2024-05-15 05:41:33.433213+00'),
(2,	'add',	'2024-05-15 05:41:36.086894+00',	'2024-05-15 05:41:36.086894+00'),
(3,	'edit',	'2024-05-15 05:41:42.740808+00',	'2024-05-15 05:41:42.740808+00'),
(4,	'delete',	'2024-05-15 05:41:46.279443+00',	'2024-05-15 05:41:46.279443+00');
