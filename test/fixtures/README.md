# Fixtures
Fixtures directory is needed for integration tests and local env setup (DB seed)

## Full demo data

The complete demo dataset is stored in
`internal/infrastructure/storage/postgres/migrations/hrms/hrms.sql`.

It starts in April 2026 and continues through May 2026 so reports, attendance,
payroll, payslips, tasks, events, notifications, and invites all have realistic
history.

All test accounts should use the same Cognito password:

```text
TestAdminPanel1!
```

The database seed creates user rows only. Passwords live in Cognito, so create
or update these Cognito users with the password above when testing login:

```text
admin@nomad-hr.test              SysAdmin
aliya.hr@nomad-hr.test           HR
daniyar.finance@nomad-hr.test    Admin
timur.lead@nomad-hr.test         Manager
aigerim.engineer@nomad-hr.test   Employee
serik.engineer@nomad-hr.test     Employee
madina.marketing@nomad-hr.test   Employee
nurlan.sales@nomad-hr.test       Employee
zhanar.ops@nomad-hr.test         Employee
```
