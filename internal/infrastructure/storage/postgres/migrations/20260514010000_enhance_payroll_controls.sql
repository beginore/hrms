-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS payroll_policies (
    id                          uuid PRIMARY KEY,
    org_id                      uuid        NOT NULL UNIQUE,
    missing_attendance_policy   varchar(40) NOT NULL DEFAULT 'PAID_REVIEW_REQUIRED',
    regular_overtime_multiplier numeric(8,6) NOT NULL DEFAULT 1.500000,
    holiday_overtime_multiplier numeric(8,6) NOT NULL DEFAULT 2.000000,
    late_penalty_mode           varchar(30) NOT NULL DEFAULT 'NONE',
    late_penalty_amount         numeric(12,2) NOT NULL DEFAULT 0,
    rounding_mode               varchar(20) NOT NULL DEFAULT 'CENT',
    vacation_pay_rate           numeric(8,6) NOT NULL DEFAULT 1.000000,
    sick_leave_pay_rate         numeric(8,6) NOT NULL DEFAULT 1.000000,
    remote_pay_rate             numeric(8,6) NOT NULL DEFAULT 1.000000,
    business_trip_pay_rate      numeric(8,6) NOT NULL DEFAULT 1.000000,
    unpaid_leave_pay_rate       numeric(8,6) NOT NULL DEFAULT 0.000000,
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE payroll_policies ADD CONSTRAINT pp_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);

ALTER TABLE payroll_items
    ADD COLUMN IF NOT EXISTS missing_days integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS review_required boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS review_reasons jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS employer_taxes_total numeric(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_employer_cost numeric(12,2) NOT NULL DEFAULT 0;

ALTER TABLE payroll_tax_rules
    ADD COLUMN IF NOT EXISTS payer varchar(20) NOT NULL DEFAULT 'EMPLOYEE';

CREATE TABLE IF NOT EXISTS payroll_corrections (
    id                 uuid PRIMARY KEY,
    org_id             uuid         NOT NULL,
    employee_id        uuid         NOT NULL,
    source_item_id     uuid,
    target_cycle_id    uuid         NOT NULL,
    adjustment_id      uuid         NOT NULL,
    amount             numeric(12,2) NOT NULL,
    type               varchar(20)  NOT NULL,
    reason             text         NOT NULL,
    created_by         uuid         NOT NULL,
    created_at         timestamptz  NOT NULL DEFAULT now()
);
ALTER TABLE payroll_corrections ADD CONSTRAINT pcorr_org_fk FOREIGN KEY (org_id) REFERENCES organizations(id);
ALTER TABLE payroll_corrections ADD CONSTRAINT pcorr_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);
ALTER TABLE payroll_corrections ADD CONSTRAINT pcorr_source_item_fk FOREIGN KEY (source_item_id) REFERENCES payroll_items(id);
ALTER TABLE payroll_corrections ADD CONSTRAINT pcorr_target_cycle_fk FOREIGN KEY (target_cycle_id) REFERENCES payroll_cycles(id);
ALTER TABLE payroll_corrections ADD CONSTRAINT pcorr_adjustment_fk FOREIGN KEY (adjustment_id) REFERENCES payroll_adjustments(id);
ALTER TABLE payroll_corrections ADD CONSTRAINT pcorr_created_by_fk FOREIGN KEY (created_by) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS payroll_corrections_employee_idx ON payroll_corrections(employee_id, target_cycle_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS payroll_corrections_employee_idx;

ALTER TABLE payroll_corrections DROP CONSTRAINT IF EXISTS pcorr_created_by_fk;
ALTER TABLE payroll_corrections DROP CONSTRAINT IF EXISTS pcorr_adjustment_fk;
ALTER TABLE payroll_corrections DROP CONSTRAINT IF EXISTS pcorr_target_cycle_fk;
ALTER TABLE payroll_corrections DROP CONSTRAINT IF EXISTS pcorr_source_item_fk;
ALTER TABLE payroll_corrections DROP CONSTRAINT IF EXISTS pcorr_employee_fk;
ALTER TABLE payroll_corrections DROP CONSTRAINT IF EXISTS pcorr_org_fk;
DROP TABLE IF EXISTS payroll_corrections;

ALTER TABLE payroll_tax_rules DROP COLUMN IF EXISTS payer;

ALTER TABLE payroll_items
    DROP COLUMN IF EXISTS total_employer_cost,
    DROP COLUMN IF EXISTS employer_taxes_total,
    DROP COLUMN IF EXISTS review_reasons,
    DROP COLUMN IF EXISTS review_required,
    DROP COLUMN IF EXISTS missing_days;

ALTER TABLE payroll_policies DROP CONSTRAINT IF EXISTS pp_org_fk;
DROP TABLE IF EXISTS payroll_policies;

-- +goose StatementEnd
