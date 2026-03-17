-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS invites
(
    id         uuid PRIMARY KEY,
    org_id     uuid NOT NULL,
    first_name varchar(30) NOT NULL,
    last_name  varchar(30) NOT NULL,
    email      varchar(255) NOT NULL,
    code       varchar(9) NOT NULL UNIQUE,
    role       varchar(100) NOT NULL DEFAULT 'Employee',
    position   varchar(100),
    expires_at timestamptz NOT NULL,
    is_used    boolean NOT NULL DEFAULT false,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE invites
    ADD CONSTRAINT invites_org_id_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE invites DROP CONSTRAINT IF EXISTS invites_org_id_fk;
DROP TABLE IF EXISTS invites;
-- +goose StatementEnd
