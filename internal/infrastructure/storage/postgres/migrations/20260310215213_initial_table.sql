-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "organizations"
(
    "id" uuid PRIMARY KEY,
    "admin_id" uuid NOT NULL UNIQUE,
    "name" varchar(50) NOT NULL,
    "vat_id" varchar(30) NOT NULL UNIQUE,
    "description" VARCHAR(250) NOT NULL,
    "address" varchar(50) NOT NULL,
    "city_id" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "users"
(
    "id" uuid PRIMARY KEY,
    "org_id" uuid NOT NULL,
    "email" VARCHAR(255) NOT NULL UNIQUE,
    "role" varchar(255) NOT NULL,
    "first_name" VARCHAR(30) NOT NULL,
    "last_name" VARCHAR(30) NOT NULL,
    "phone_number" VARCHAR(20) NOT NULL UNIQUE,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'Unverified',
    "created_at" timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "consents"

(
    "id" uuid PRIMARY KEY,
    "user_id" uuid NOT NULL,
    "org_id" uuid NOT NULL,
    "document_type" varchar(20) NOT NULL,
    "version" varchar(20) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE consents ADD CONSTRAINT consents_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE consents ADD CONSTRAINT consents_org_id_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE consents DROP CONSTRAINT IF EXISTS consents_user_id_fk;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_org_fk;

DROP TABLE IF EXISTS consents;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd
