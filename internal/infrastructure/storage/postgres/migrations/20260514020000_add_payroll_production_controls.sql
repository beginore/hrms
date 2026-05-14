-- +goose Up
-- +goose StatementBegin

ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS hire_date date,
    ADD COLUMN IF NOT EXISTS termination_date date;

ALTER TABLE payroll_cycles
    ADD COLUMN IF NOT EXISTS currency varchar(3) NOT NULL DEFAULT 'KZT';

ALTER TABLE payroll_items
    ADD COLUMN IF NOT EXISTS currency varchar(3) NOT NULL DEFAULT 'KZT',
    ADD COLUMN IF NOT EXISTS reviewed_by uuid,
    ADD COLUMN IF NOT EXISTS reviewed_at timestamptz,
    ADD COLUMN IF NOT EXISTS review_comment text;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'pi_reviewed_by_fk') THEN
        ALTER TABLE payroll_items ADD CONSTRAINT pi_reviewed_by_fk FOREIGN KEY (reviewed_by) REFERENCES users(id);
    END IF;
END $$;

ALTER TABLE payroll_adjustments
    ADD COLUMN IF NOT EXISTS currency varchar(3) NOT NULL DEFAULT 'KZT';

ALTER TABLE payroll_corrections
    ADD COLUMN IF NOT EXISTS currency varchar(3) NOT NULL DEFAULT 'KZT';

CREATE TABLE IF NOT EXISTS salary_history (
    id             uuid PRIMARY KEY,
    org_id         uuid         NOT NULL,
    employee_id    uuid         NOT NULL,
    salary_rate    numeric(12,2) NOT NULL,
    effective_from date         NOT NULL,
    effective_to   date,
    created_by     uuid         NOT NULL,
    created_at     timestamptz  NOT NULL DEFAULT now()
);
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sh_org_fk') THEN
        ALTER TABLE salary_history ADD CONSTRAINT sh_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sh_employee_fk') THEN
        ALTER TABLE salary_history ADD CONSTRAINT sh_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sh_created_by_fk') THEN
        ALTER TABLE salary_history ADD CONSTRAINT sh_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id);
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS salary_history_employee_period_idx ON salary_history(employee_id, effective_from, effective_to);

CREATE TABLE IF NOT EXISTS payroll_attendance_overrides (
    id               uuid PRIMARY KEY,
    org_id           uuid        NOT NULL,
    employee_id      uuid        NOT NULL,
    date             date        NOT NULL,
    paid_day         boolean     NOT NULL,
    attendance_type  varchar(30) NOT NULL DEFAULT 'MANUAL',
    attendance_status varchar(30) NOT NULL DEFAULT 'PRESENT',
    overtime_minutes integer     NOT NULL DEFAULT 0,
    note             text,
    created_by       uuid        NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (employee_id, date)
);
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'pao_org_fk') THEN
        ALTER TABLE payroll_attendance_overrides ADD CONSTRAINT pao_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'pao_employee_fk') THEN
        ALTER TABLE payroll_attendance_overrides ADD CONSTRAINT pao_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'pao_created_by_fk') THEN
        ALTER TABLE payroll_attendance_overrides ADD CONSTRAINT pao_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id);
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS payroll_attendance_overrides_employee_date_idx ON payroll_attendance_overrides(employee_id, date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS payroll_attendance_overrides_employee_date_idx;
ALTER TABLE payroll_attendance_overrides DROP CONSTRAINT IF EXISTS pao_created_by_fk;
ALTER TABLE payroll_attendance_overrides DROP CONSTRAINT IF EXISTS pao_employee_fk;
ALTER TABLE payroll_attendance_overrides DROP CONSTRAINT IF EXISTS pao_org_fk;
DROP TABLE IF EXISTS payroll_attendance_overrides;

DROP INDEX IF EXISTS salary_history_employee_period_idx;
ALTER TABLE salary_history DROP CONSTRAINT IF EXISTS sh_created_by_fk;
ALTER TABLE salary_history DROP CONSTRAINT IF EXISTS sh_employee_fk;
ALTER TABLE salary_history DROP CONSTRAINT IF EXISTS sh_org_fk;
DROP TABLE IF EXISTS salary_history;

ALTER TABLE payroll_corrections DROP COLUMN IF EXISTS currency;
ALTER TABLE payroll_adjustments DROP COLUMN IF EXISTS currency;

ALTER TABLE payroll_items DROP CONSTRAINT IF EXISTS pi_reviewed_by_fk;
ALTER TABLE payroll_items
    DROP COLUMN IF EXISTS review_comment,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS currency;

ALTER TABLE payroll_cycles DROP COLUMN IF EXISTS currency;

ALTER TABLE employees
    DROP COLUMN IF EXISTS termination_date,
    DROP COLUMN IF EXISTS hire_date;

-- +goose StatementEnd
