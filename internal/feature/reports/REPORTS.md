# Reports & Analytics Module

The Reports & Analytics module provides read-only analytics for dashboard cards, charts, tables, and downloadable reports.

It does not recalculate payroll, attendance, or employee data. It reads existing records from other modules and returns aggregated data for frontend visualization.

## Purpose

This module supports:

- payroll reports;
- attendance reports;
- employee statistics;
- department statistics;
- dashboard summary cards;
- chart-ready API responses;
- CSV and PDF report export.

## Access Rules

Reports are available only to organisation-level management roles:

- `SysAdmin`
- `Admin`
- `HR`

Employees cannot access organisation-wide reports.

## Data Sources

Reports are built from existing tables:

- `users`
- `employees`
- `departments`
- `attendance_records`
- `leave_requests`
- `payroll_cycles`
- `payroll_items`

No new database tables are required for this module.

## Important Payroll Rule

Payroll reports include only payroll cycles with final or management-approved statuses:

- `APPROVED`
- `READY_FOR_PAYMENT`
- `PAID`

Draft payroll cycles are excluded from reports because they can contain unfinished or review-required calculations.

## Endpoints

All endpoints require:

```http
Authorization: Bearer ACCESS_TOKEN
```

### `GET /v1/reports/dashboard`

Returns dashboard summary metrics for top-level cards.

Includes:

- total employees;
- active employees;
- present today;
- late today;
- on leave today;
- absent today;
- current month payroll;
- employee growth for current month;
- payroll growth percent;
- attendance rate.

Example response:

```json
{
  "total_employees": 2,
  "active_employees": 2,
  "present_today": 0,
  "late_today": 0,
  "on_leave_today": 0,
  "absent_today": 2,
  "monthly_payroll": "241999.96",
  "payroll_currency": "KZT",
  "employee_growth": 2,
  "payroll_growth_percent": "100.00",
  "attendance_rate": "0.00"
}
```

### `GET /v1/reports/payroll`

Returns payroll summary for a period.

Query parameters:

```text
period_start=YYYY-MM-DD
period_end=YYYY-MM-DD
```

If no period is provided, the current month is used.

Example:

```http
GET /v1/reports/payroll?period_start=2026-05-01&period_end=2026-05-31
```

Includes:

- cycles count;
- employees paid;
- gross salary total;
- net salary total;
- bonuses total;
- deductions total;
- taxes total;
- employer taxes total;
- total employer cost;
- average net salary.

### `GET /v1/reports/payroll/trends`

Returns monthly payroll totals as chart-ready data.

Example:

```http
GET /v1/reports/payroll/trends?period_start=2026-01-01&period_end=2026-12-31
```

Example response:

```json
{
  "labels": ["2026-01", "2026-02", "2026-03"],
  "series": [
    {
      "name": "Net payroll",
      "data": [250000, 275000, 310000]
    }
  ]
}
```

Frontend can render this as a line chart or bar chart.

### `GET /v1/reports/payroll/departments`

Returns department payroll table and chart data.

Example:

```http
GET /v1/reports/payroll/departments?period_start=2026-05-01&period_end=2026-05-31
```

Includes per department:

- employees count;
- average base salary;
- gross salary total;
- net salary total;
- bonuses total;
- deductions total;
- taxes total;
- employer taxes total;
- total employer cost.

Note:

`average_salary` is the average employee salary rate from `employees.salary_rate`. It is not the paid payroll amount.

### `GET /v1/reports/attendance/today`

Returns attendance breakdown for one day.

Query parameter:

```text
date=YYYY-MM-DD
```

If no date is provided, today is used.

Example:

```http
GET /v1/reports/attendance/today?date=2026-05-15
```

Includes:

- active employees;
- present;
- late;
- on leave;
- absent;
- remote;
- attendance rate;
- pie-chart-ready items.

### `GET /v1/reports/attendance/weekly`

Returns 7-day attendance chart data.

Query parameter:

```text
week_start=YYYY-MM-DD
```

Example:

```http
GET /v1/reports/attendance/weekly?week_start=2026-05-11
```

Returns:

- labels for seven days;
- present series;
- late series;
- on leave series;
- absent series.

### `GET /v1/reports/employees/statistics`

Returns employee statistics.

Includes:

- total employees;
- active employees;
- inactive employees;
- average salary;
- employees by department;
- employees by role;
- employees by status.

The response is ready for bar and pie charts.

### `GET /v1/reports/departments/statistics`

Returns department analytics for a period.

Example:

```http
GET /v1/reports/departments/statistics?period_start=2026-05-01&period_end=2026-05-31
```

Includes per department:

- employees count;
- average base salary;
- attendance rate;
- absent days;
- late days;
- overtime minutes;
- net salary total;
- total employer cost.

## Export Endpoints

There are two explicit export endpoints to avoid Swagger UI download confusion.

### `GET /v1/reports/export/csv`

Exports payroll or attendance report as CSV.

Query parameters:

```text
type=payroll | attendance
period_start=YYYY-MM-DD
period_end=YYYY-MM-DD
```

Examples:

```http
GET /v1/reports/export/csv?type=payroll&period_start=2026-05-01&period_end=2026-05-31
```

```http
GET /v1/reports/export/csv?type=attendance&period_start=2026-05-01&period_end=2026-05-31
```

### `GET /v1/reports/export/pdf`

Exports payroll or attendance report as PDF.

Examples:

```http
GET /v1/reports/export/pdf?type=payroll&period_start=2026-05-01&period_end=2026-05-31
```

```http
GET /v1/reports/export/pdf?type=attendance&period_start=2026-05-01&period_end=2026-05-31
```

The response is binary:

```http
Content-Type: application/pdf
Content-Disposition: attachment; filename="payroll-report-2026-05-01-2026-05-31.pdf"
```

Frontend should handle this response as a file download, not as JSON.

## Frontend Usage

The frontend owns the visual layout:

- cards;
- charts;
- icons;
- colors;
- tables;
- buttons.

The backend returns chart-ready JSON.

Bar chart shape:

```json
{
  "labels": ["Engineering", "Human Resources"],
  "series": [
    {
      "name": "Employees",
      "data": [1, 1]
    }
  ]
}
```

Pie chart shape:

```json
{
  "items": [
    { "label": "Present", "value": 236 },
    { "label": "Late", "value": 8 },
    { "label": "Absent", "value": 4 }
  ]
}
```

Line chart shape:

```json
{
  "labels": ["Mon 11", "Tue 12", "Wed 13"],
  "series": [
    {
      "name": "Present",
      "data": [20, 22, 21]
    },
    {
      "name": "Late",
      "data": [2, 1, 3]
    }
  ]
}
```

## Swagger Testing Flow

1. Start the backend.
2. Open:

```http
http://localhost:8080/swagger/index.html
```

3. Click `Authorize`.
4. Enter:

```text
Bearer ACCESS_TOKEN
```

5. Test these first:

```http
GET /v1/reports/dashboard
GET /v1/reports/employees/statistics
GET /v1/reports/attendance/today
```

6. If payroll data is empty, create/calculate/approve or prepare a payroll cycle first.

7. Test exports:

```http
GET /v1/reports/export/pdf?type=payroll&period_start=2026-05-01&period_end=2026-05-31
GET /v1/reports/export/csv?type=payroll&period_start=2026-05-01&period_end=2026-05-31
```

## Common Notes

### Why payroll summary can be zero

Payroll reports exclude draft cycles.

If the response shows:

```text
Cycles: 0
Employees paid: 0
Net salary: 0
```

then there are no `APPROVED`, `READY_FOR_PAYMENT`, or `PAID` cycles in the selected period.

### Why average salary can be non-zero when payroll is zero

Department average salary comes from `employees.salary_rate`.

This is the employee base salary rate, not the paid payroll total.

Example:

```text
Human Resources | 1 | 250000.00 KZT | 0 KZT | 0 KZT
```

means:

- one employee exists in Human Resources;
- their average base salary is `250000.00 KZT`;
- no payroll was paid for the selected period.

### Why absent today can equal active employees

If active employees have no attendance records and no approved leave for today, the dashboard treats them as absent.

This is useful for analytics, but for production accuracy attendance data should be generated daily through manual check-in/check-out, SKUD events, or scheduled attendance processing.
