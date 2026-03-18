# Notification Policy

This module is backend-driven. End users do not create notifications through public HTTP endpoints.

Roles:
- `SysAdmin`: organization creator / super admin in current codebase
- `Admin`: department or operational manager
- `Employee`: regular employee

Delivery rules:
- `system`
  - created by backend services for operational or account lifecycle events
  - targets one user or a role audience
- `payroll`
  - created by payroll service
  - typical audience: `Admin` for review/approval, then `Employee` when payroll is available
- `salary`
  - created by salary payment flow
  - typical audience: one `Employee`

Public API:
- `GET /v1/notifications`
- `PATCH /v1/notifications/:id/read`
- `PATCH /v1/notifications/read-all`

Internal service entrypoints:
- `NotifySystemToUser`
- `NotifySystemToRole`
- `NotifyPayrollToAdmins`
- `NotifySalaryToEmployee`
