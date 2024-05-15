-- Create TRIGGER function for auto update updated_at column after records being altered
CREATE OR REPLACE FUNCTION update_updated_at_column()   
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;   
END;
$$ language 'plpgsql';

-- Create resources table
DROP SEQUENCE IF EXISTS resources_id_seq;
CREATE SEQUENCE resources_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE "public"."resources" (
    "id" integer DEFAULT nextval('resources_id_seq') NOT NULL,
    "name" character varying(255) NOT NULL,
    "disabled" boolean DEFAULT false NOT NULL,
    "created_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "resources_pkey" PRIMARY KEY ("id")
) WITH (oids = false);

-- Apply trigger update_updated_at_column()
CREATE TRIGGER update_resources_updated_at BEFORE UPDATE ON resources FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Pre-populate some default resources
INSERT INTO "resources" ("id", "name", "disabled", "created_at", "updated_at") VALUES
(1,	'users',	'0',	'2024-05-14 06:54:25.780889+00',	'2024-05-14 06:54:25.780889+00'),
(2,	'groups',	'0',	'2024-05-14 06:54:30.014063+00',	'2024-05-14 06:54:30.014063+00'),
(3,	'todos',	'0',	'2024-05-14 06:54:33.907284+00',	'2024-05-14 06:54:33.907284+00');
