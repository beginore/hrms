-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS documents
(
    "id"         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "type"       varchar(30) NOT NULL CHECK (type IN ('PRIVACY_POLICY', 'TERMS_AND_CONDITIONS')),
    "version"    varchar(20) NOT NULL,
    "url"        text NOT NULL,
    "is_active"  boolean NOT NULL DEFAULT true,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    UNIQUE ("type", "version")
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS documents;
-- +goose StatementEnd