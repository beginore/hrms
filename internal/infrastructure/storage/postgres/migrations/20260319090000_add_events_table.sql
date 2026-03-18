-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events
(
    id              uuid PRIMARY KEY,
    title           varchar(150) NOT NULL,
    description     text         NOT NULL,
    starts_at       timestamptz  NOT NULL,
    ends_at         timestamptz  NOT NULL,
    scope           varchar(20)  NOT NULL,
    department_id   uuid,
    created_by      uuid         NOT NULL,
    created_by_role varchar(50)  NOT NULL,
    organization_id uuid         NOT NULL,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT events_scope_check CHECK (scope IN ('global', 'department')),
    CONSTRAINT events_department_scope_check CHECK (
        (scope = 'global' AND department_id IS NULL) OR
        (scope = 'department' AND department_id IS NOT NULL)
    )
);

ALTER TABLE events
    ADD CONSTRAINT events_department_id_fk FOREIGN KEY (department_id) REFERENCES departments(id);
ALTER TABLE events
    ADD CONSTRAINT events_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id);
ALTER TABLE events
    ADD CONSTRAINT events_organization_id_fk FOREIGN KEY (organization_id) REFERENCES organizations(id);

CREATE INDEX IF NOT EXISTS idx_events_org_starts_at ON events (organization_id, starts_at);
CREATE INDEX IF NOT EXISTS idx_events_department_starts_at ON events (department_id, starts_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_events_department_starts_at;
DROP INDEX IF EXISTS idx_events_org_starts_at;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_organization_id_fk;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_created_by_fk;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_department_id_fk;
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
