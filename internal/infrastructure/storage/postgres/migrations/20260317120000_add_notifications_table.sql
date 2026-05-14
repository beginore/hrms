-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notifications
(
    id         uuid PRIMARY KEY,
    user_id    uuid         NOT NULL,
    org_id     uuid,
    type       varchar(20)  NOT NULL,
    title      varchar(120) NOT NULL,
    message    text         NOT NULL,
    metadata   jsonb        NOT NULL DEFAULT '{}'::jsonb,
    is_read    boolean      NOT NULL DEFAULT false,
    read_at    timestamptz,
    created_at timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT notifications_type_check CHECK (type IN ('payroll', 'salary', 'system'))
);

ALTER TABLE notifications
    ADD CONSTRAINT notifications_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE notifications
    ADD CONSTRAINT notifications_org_id_fk FOREIGN KEY (org_id) REFERENCES organizations(id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created_at
    ON notifications (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_user_is_read
    ON notifications (user_id, is_read);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notifications_user_is_read;
DROP INDEX IF EXISTS idx_notifications_user_created_at;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_org_id_fk;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_user_id_fk;
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
