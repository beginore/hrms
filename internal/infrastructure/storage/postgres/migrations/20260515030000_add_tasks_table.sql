-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS tasks (
    id           uuid         PRIMARY KEY,
    org_id       uuid         NOT NULL,
    created_by   uuid         NOT NULL,
    assigned_to  uuid         NOT NULL,
    title        varchar(255) NOT NULL,
    description  text,
    due_date     date         NOT NULL,
    status       varchar(20)  NOT NULL DEFAULT 'PENDING', -- PENDING | APPROVED | REJECTED
    reviewed_by  uuid,
    reviewed_at  timestamptz,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now()
);
ALTER TABLE tasks ADD CONSTRAINT tasks_org_fk          FOREIGN KEY (org_id)      REFERENCES organizations(id);
ALTER TABLE tasks ADD CONSTRAINT tasks_created_by_fk   FOREIGN KEY (created_by)  REFERENCES employees(id);
ALTER TABLE tasks ADD CONSTRAINT tasks_assigned_to_fk  FOREIGN KEY (assigned_to) REFERENCES employees(id);
ALTER TABLE tasks ADD CONSTRAINT tasks_reviewed_by_fk  FOREIGN KEY (reviewed_by) REFERENCES employees(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
-- +goose StatementEnd
