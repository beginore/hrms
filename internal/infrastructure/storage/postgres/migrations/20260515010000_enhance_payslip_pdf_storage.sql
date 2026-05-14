-- +goose Up
-- +goose StatementBegin

ALTER TABLE payslips
    ADD COLUMN IF NOT EXISTS pdf_content bytea,
    ADD COLUMN IF NOT EXISTS pdf_filename varchar(255),
    ADD COLUMN IF NOT EXISTS pdf_generated_at timestamptz,
    ADD COLUMN IF NOT EXISTS pdf_sha256 varchar(64);

ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_payroll_item_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS payslips_active_payroll_item_idx
    ON payslips(payroll_item_id)
    WHERE status <> 'VOID';

CREATE INDEX IF NOT EXISTS payslips_pdf_generated_idx
    ON payslips(org_id, pdf_generated_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS payslips_pdf_generated_idx;
DROP INDEX IF EXISTS payslips_active_payroll_item_idx;

ALTER TABLE payslips DROP COLUMN IF EXISTS pdf_sha256;
ALTER TABLE payslips DROP COLUMN IF EXISTS pdf_generated_at;
ALTER TABLE payslips DROP COLUMN IF EXISTS pdf_filename;
ALTER TABLE payslips DROP COLUMN IF EXISTS pdf_content;

-- +goose StatementEnd
