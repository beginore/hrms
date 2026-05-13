-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS work_schedules (
    id            uuid PRIMARY KEY,
    org_id        uuid        NOT NULL,
    employee_id   uuid        NOT NULL UNIQUE,
    work_start    time        NOT NULL DEFAULT '09:00:00',
    work_end      time        NOT NULL DEFAULT '18:00:00',
    late_threshold_minutes integer NOT NULL DEFAULT 15,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE work_schedules ADD CONSTRAINT ws_org_fk      FOREIGN KEY (org_id)      REFERENCES organizations(id);
ALTER TABLE work_schedules ADD CONSTRAINT ws_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);

CREATE TABLE IF NOT EXISTS skud_events (
    id           uuid        PRIMARY KEY,
    org_id       uuid        NOT NULL,
    employee_id  uuid        NOT NULL,
    event_type   varchar(10) NOT NULL, -- ENTER | EXIT
    device_id    varchar(100),
    occurred_at  timestamptz NOT NULL,
    processed    boolean     NOT NULL DEFAULT false,
    created_at   timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE skud_events ADD CONSTRAINT se_org_fk      FOREIGN KEY (org_id)      REFERENCES organizations(id);
ALTER TABLE skud_events ADD CONSTRAINT se_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);

CREATE TABLE IF NOT EXISTS leave_requests (
    id           uuid         PRIMARY KEY,
    org_id       uuid         NOT NULL,
    employee_id  uuid         NOT NULL,
    type         varchar(20)  NOT NULL, -- SICK_LEAVE | VACATION | REMOTE | BUSINESS_TRIP | UNPAID_LEAVE
    start_date   date         NOT NULL,
    end_date     date         NOT NULL,
    reason       text,
    document_url text,
    status       varchar(20)  NOT NULL DEFAULT 'PENDING', -- PENDING | APPROVED | REJECTED
    reviewed_by  uuid,
    reviewed_at  timestamptz,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now()
);
ALTER TABLE leave_requests ADD CONSTRAINT lr_org_fk         FOREIGN KEY (org_id)      REFERENCES organizations(id);
ALTER TABLE leave_requests ADD CONSTRAINT lr_employee_fk    FOREIGN KEY (employee_id) REFERENCES employees(id);
ALTER TABLE leave_requests ADD CONSTRAINT lr_reviewed_by_fk FOREIGN KEY (reviewed_by) REFERENCES employees(id);

CREATE TABLE IF NOT EXISTS attendance_records (
    id           uuid         PRIMARY KEY,
    org_id       uuid         NOT NULL,
    employee_id  uuid         NOT NULL,
    date         date         NOT NULL,
    type         varchar(20)  NOT NULL, -- OFFICE | REMOTE | SICK_LEAVE | VACATION | BUSINESS_TRIP | ABSENT
    source       varchar(10)  NOT NULL, -- SKUD | MANUAL | SYSTEM
    status       varchar(20)  NOT NULL, -- PRESENT | LATE | ABSENT | ON_LEAVE
    check_in     timestamptz,
    check_out    timestamptz,
    note         text,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now(),
    UNIQUE (employee_id, date)
);
ALTER TABLE attendance_records ADD CONSTRAINT ar_org_fk      FOREIGN KEY (org_id)      REFERENCES organizations(id);
ALTER TABLE attendance_records ADD CONSTRAINT ar_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE attendance_records DROP CONSTRAINT IF EXISTS ar_employee_fk;
ALTER TABLE attendance_records DROP CONSTRAINT IF EXISTS ar_org_fk;
DROP TABLE IF EXISTS attendance_records;

ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS lr_reviewed_by_fk;
ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS lr_employee_fk;
ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS lr_org_fk;
DROP TABLE IF EXISTS leave_requests;

ALTER TABLE skud_events DROP CONSTRAINT IF EXISTS se_employee_fk;
ALTER TABLE skud_events DROP CONSTRAINT IF EXISTS se_org_fk;
DROP TABLE IF EXISTS skud_events;

ALTER TABLE work_schedules DROP CONSTRAINT IF EXISTS ws_employee_fk;
ALTER TABLE work_schedules DROP CONSTRAINT IF EXISTS ws_org_fk;
DROP TABLE IF EXISTS work_schedules;

-- +goose StatementEnd