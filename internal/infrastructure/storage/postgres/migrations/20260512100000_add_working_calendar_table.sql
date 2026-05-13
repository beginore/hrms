-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS working_calendar (
    id         uuid         PRIMARY KEY,
    org_id     uuid         NOT NULL,
    date       date         NOT NULL,
    type       varchar(20)  NOT NULL, -- HOLIDAY | WORKDAY
    name       varchar(100),
    created_at timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (org_id, date)
);
ALTER TABLE working_calendar ADD CONSTRAINT wc_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE working_calendar DROP CONSTRAINT IF EXISTS wc_org_fk;
DROP TABLE IF EXISTS working_calendar;

-- +goose StatementEnd
