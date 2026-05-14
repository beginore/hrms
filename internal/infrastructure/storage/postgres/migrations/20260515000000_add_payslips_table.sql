-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS payslips (
    id                    uuid PRIMARY KEY,
    org_id                uuid         NOT NULL,
    employee_id           uuid         NOT NULL,
    payroll_cycle_id      uuid         NOT NULL,
    payroll_item_id       uuid         NOT NULL UNIQUE,
    period_start          date         NOT NULL,
    period_end            date         NOT NULL,
    status                varchar(20)  NOT NULL DEFAULT 'GENERATED',
    currency              varchar(3)   NOT NULL DEFAULT 'KZT',
    base_salary           numeric(12,2) NOT NULL,
    overtime_amount       numeric(12,2) NOT NULL DEFAULT 0,
    bonuses_total         numeric(12,2) NOT NULL DEFAULT 0,
    deductions_total      numeric(12,2) NOT NULL DEFAULT 0,
    taxes_total           numeric(12,2) NOT NULL DEFAULT 0,
    gross_salary          numeric(12,2) NOT NULL,
    net_salary            numeric(12,2) NOT NULL,
    employer_taxes_total  numeric(12,2) NOT NULL DEFAULT 0,
    total_employer_cost   numeric(12,2) NOT NULL DEFAULT 0,
    payload_snapshot      jsonb        NOT NULL,
    sent_to_email         varchar(255),
    sent_at               timestamptz,
    generated_by          uuid         NOT NULL,
    generated_at          timestamptz  NOT NULL DEFAULT now(),
    voided_by             uuid,
    voided_at             timestamptz,
    void_reason           text,
    created_at            timestamptz  NOT NULL DEFAULT now(),
    updated_at            timestamptz  NOT NULL DEFAULT now()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payslips_org_fk') THEN
        ALTER TABLE payslips ADD CONSTRAINT payslips_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payslips_employee_fk') THEN
        ALTER TABLE payslips ADD CONSTRAINT payslips_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payslips_cycle_fk') THEN
        ALTER TABLE payslips ADD CONSTRAINT payslips_cycle_fk FOREIGN KEY (payroll_cycle_id) REFERENCES payroll_cycles(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payslips_item_fk') THEN
        ALTER TABLE payslips ADD CONSTRAINT payslips_item_fk FOREIGN KEY (payroll_item_id) REFERENCES payroll_items(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payslips_generated_by_fk') THEN
        ALTER TABLE payslips ADD CONSTRAINT payslips_generated_by_fk FOREIGN KEY (generated_by) REFERENCES users(id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payslips_voided_by_fk') THEN
        ALTER TABLE payslips ADD CONSTRAINT payslips_voided_by_fk FOREIGN KEY (voided_by) REFERENCES users(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS payslips_org_cycle_idx ON payslips(org_id, payroll_cycle_id);
CREATE INDEX IF NOT EXISTS payslips_employee_period_idx ON payslips(employee_id, period_start, period_end);
CREATE INDEX IF NOT EXISTS payslips_status_idx ON payslips(org_id, status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS payslips_status_idx;
DROP INDEX IF EXISTS payslips_employee_period_idx;
DROP INDEX IF EXISTS payslips_org_cycle_idx;

ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_voided_by_fk;
ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_generated_by_fk;
ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_item_fk;
ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_cycle_fk;
ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_employee_fk;
ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_org_fk;
DROP TABLE IF EXISTS payslips;

-- +goose StatementEnd
