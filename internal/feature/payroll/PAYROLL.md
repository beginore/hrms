# Payroll Module README

Payroll module is responsible for automated salary calculation, payroll cycle management, bonuses, deductions, overtime, taxes, attendance impact, review flow, and payment preparation without real bank transfers.

The module does not send money. It calculates payroll, prepares a payment batch, exports data for accounting or bank upload, and allows HR/Admin to mark the cycle as paid after external payment is completed.

## Main Formula

```text
gross_salary =
base_salary
+ overtime_amount
+ bonuses_total

net_salary =
gross_salary
- deductions_total
- employee_taxes

total_employer_cost =
gross_salary
+ employer_taxes
```

Payroll calculation also considers:

- employee salary and salary history;
- attendance records;
- missing attendance;
- approved leave requests;
- work schedule;
- overtime;
- bonuses;
- deductions;
- tax rules;
- payroll policy;
- hire date and termination date;
- manual attendance overrides;
- review-required flags.

## Payroll Cycle Status Flow

```text
DRAFT
-> CALCULATED
-> APPROVED
-> READY_FOR_PAYMENT
-> PAID
```

### Status Meaning

| Status | Meaning |
| --- | --- |
| `DRAFT` | Cycle was created but payroll items were not calculated yet. |
| `CALCULATED` | Payroll items were calculated and can still be reviewed or recalculated. |
| `APPROVED` | Payroll was approved by HR/Admin and is locked from changes. |
| `READY_FOR_PAYMENT` | Payroll is prepared for external payment and can be exported as CSV. |
| `PAID` | HR/Admin marked the cycle as paid after payment was completed outside the system. |

## Recommended Testing Flow

```text
1. Create or configure payroll policy
2. Create tax rules or KZ preset
3. Create salary history if needed
4. Add attendance overrides if needed
5. Create payroll cycle
6. Preview cycle
7. Calculate cycle
8. Review items with review_required=true
9. Approve cycle
10. Prepare payment
11. Export payment batch CSV
12. Mark cycle as paid
```

## Endpoints

All payroll endpoints are protected and require:

```http
Authorization: Bearer <access_token>
```

Base path:

```text
/v1
```

## Payroll Preview

### `POST /payroll/preview`

Calculates payroll for one employee without saving anything to the database.

Use this endpoint when HR wants to check how salary will be calculated before creating or recalculating a real payroll cycle.

Main function:

- quick one-employee salary preview;
- safe testing of salary, attendance, taxes, and overtime;
- does not create payroll item;
- does not change payroll cycle status.

Example body:

```json
{
  "employee_id": "EMPLOYEE_ID_HERE",
  "period_start": "2026-05-01",
  "period_end": "2026-05-31"
}
```

## Payroll Cycles

### `POST /payroll/cycles`

Creates a payroll cycle for an organization and period.

A cycle is a payroll period, usually one month. For example, May 2026 payroll.

Main function:

- creates payroll period;
- stores period start and end;
- sets default status `DRAFT`;
- sets currency, usually `KZT`.

Example body:

```json
{
  "period_start": "2026-05-01",
  "period_end": "2026-05-31",
  "currency": "KZT"
}
```

### `GET /payroll/cycles`

Returns all payroll cycles for the caller's organization.

Main function:

- show payroll history;
- check current cycle status;
- find cycle IDs for testing;
- track `DRAFT`, `CALCULATED`, `APPROVED`, `READY_FOR_PAYMENT`, and `PAID` cycles.

### `GET /payroll/cycles/{id}`

Returns one payroll cycle by ID.

Main function:

- inspect payroll cycle details;
- check period, currency, status, approval info, paid timestamp;
- verify current status before the next action.

### `POST /payroll/cycles/{id}/preview`

Calculates all employees in the cycle without saving payroll items.

Main function:

- preview whole cycle before real calculation;
- check expected salary results;
- detect possible review-required items early;
- does not update cycle status.

### `PATCH /payroll/cycles/{id}/calculate`

Calculates and saves payroll items for all active employees in the organization.

Main function:

- calculates salary for all employees;
- saves payroll items;
- applies attendance, overtime, bonuses, deductions, taxes, salary history, and policy;
- sets cycle and items to `CALCULATED`;
- creates calculation snapshots for audit/debugging.

Example body:

```json
{
  "recalculate": true
}
```

### `GET /payroll/cycles/{id}/items`

Returns all calculated payroll items for a cycle.

Main function:

- inspect each employee's calculated salary;
- check `review_required`;
- check `net_salary`, `gross_salary`, taxes, deductions, overtime;
- get payroll item IDs for review.

### `PATCH /payroll/cycles/{id}/approve`

Approves a calculated payroll cycle.

Main function:

- locks calculated payroll for payment preparation;
- moves cycle from `CALCULATED` to `APPROVED`;
- blocks further recalculation and changes;
- requires all items to have `review_required=false`.

Important:

If at least one payroll item has:

```json
"review_required": true
```

then approval will fail. Review those items first.

### `PATCH /payroll/cycles/{id}/prepare-payment`

Moves approved payroll to payment preparation state.

Main function:

- moves cycle from `APPROVED` to `READY_FOR_PAYMENT`;
- means payroll is ready for external payment;
- allows CSV export;
- does not send real money.

This is the correct replacement for real bank transfer in this HRMS project.

### `GET /payroll/cycles/{id}/export`

Exports payroll cycle as CSV payment batch.

Main function:

- creates CSV for accounting or bank upload;
- includes employee ID, period, gross salary, deductions, taxes, net salary, currency, status;
- available only when cycle is `READY_FOR_PAYMENT` or `PAID`.

This endpoint prepares payment data but does not execute payment.

### `PATCH /payroll/cycles/{id}/mark-paid`

Marks payroll cycle as paid after external payment is completed.

Main function:

- moves cycle from `READY_FOR_PAYMENT` to `PAID`;
- saves paid timestamp;
- marks payroll items as `PAID`;
- confirms payment was completed outside the system.

Important:

This endpoint does not transfer money. It only records payment completion.

### `PATCH /payroll/cycles/{id}/reopen`

Reopens an approved or ready-for-payment cycle back to calculated state.

Main function:

- moves cycle from `APPROVED` or `READY_FOR_PAYMENT` back to `CALCULATED`;
- allows HR/Admin to fix payroll before payment;
- cannot reopen `PAID` cycles.

Use this if HR found a mistake before payment was completed.

## Payroll Items

### `GET /payroll/items/{id}`

Returns one payroll item.

Main function:

- inspect one employee's payroll calculation;
- see full calculation snapshot;
- check attendance impact, bonuses, deductions, taxes, and review reasons.

Employees can view their own item. Payroll managers can view items in their organization.

### `PATCH /payroll/items/{id}/review`

Marks payroll item as reviewed.

Main function:

- clears `review_required`;
- saves reviewer ID and comment;
- allows cycle approval after all problematic items are reviewed.

Use this endpoint only when payroll item has:

```json
"review_required": true
```

Example body:

```json
{
  "comment": "Checked missing attendance manually"
}
```

## Payroll Adjustments

### `POST /payroll/adjustments`

Creates a bonus or deduction for an employee.

Main function:

- add bonus;
- add deduction;
- define whether bonus is taxable;
- attach adjustment to a payroll cycle;
- include reason and category.

Example bonus:

```json
{
  "employee_id": "EMPLOYEE_ID_HERE",
  "cycle_id": "CYCLE_ID_HERE",
  "type": "BONUS",
  "category": "KPI",
  "amount": "30000.00",
  "currency": "KZT",
  "is_taxable": true,
  "reason": "Monthly KPI bonus"
}
```

Example deduction:

```json
{
  "employee_id": "EMPLOYEE_ID_HERE",
  "cycle_id": "CYCLE_ID_HERE",
  "type": "DEDUCTION",
  "category": "ADVANCE_PAYMENT",
  "amount": "10000.00",
  "currency": "KZT",
  "is_taxable": false,
  "reason": "Advance payment deduction"
}
```

### `GET /payroll/adjustments`

Returns payroll adjustments.

Main function:

- list bonuses and deductions;
- filter by cycle;
- filter by employee;
- check what will affect payroll calculation.

Query params:

```text
cycle_id
employee_id
```

### `DELETE /payroll/adjustments/{id}`

Deletes payroll adjustment.

Main function:

- remove incorrect bonus or deduction;
- available while related cycle is still editable;
- useful before recalculation.

## Tax Rules

### `POST /payroll/tax-rules`

Creates a custom payroll tax rule.

Main function:

- configure employee taxes;
- configure employer taxes;
- define tax rate;
- define taxable base;
- define effective period;
- define thresholds.

Supported `applies_to` values:

```text
GROSS
TAXABLE_INCOME
BASE_ONLY
```

Supported `payer` values:

```text
EMPLOYEE
EMPLOYER
```

Example body:

```json
{
  "country": "KZ",
  "name": "Individual Income Tax",
  "rate": "0.100000",
  "applies_to": "TAXABLE_INCOME",
  "payer": "EMPLOYEE",
  "effective_from": "2026-01-01"
}
```

### `POST /payroll/tax-rules/kz-preset`

Creates basic Kazakhstan payroll tax preset for the organization.

Main function:

- quickly add default KZ-style rules for testing;
- creates employee income tax and employer tax rules;
- useful for Swagger testing.

Important:

This preset is simplified. For production, Kazakhstan payroll taxes must be legally verified and expanded with limits, exemptions, and official rules.

### `GET /payroll/tax-rules`

Returns active payroll tax rules for the organization.

Main function:

- inspect configured tax rules;
- confirm KZ preset was created;
- verify rates and payer type.

## Payroll Policy

### `GET /payroll/policy`

Returns payroll calculation policy for the organization.

Main function:

- check missing attendance policy;
- check overtime multipliers;
- check late penalty settings;
- check rounding mode;
- check leave pay rates.

### `PUT /payroll/policy`

Creates or updates payroll calculation policy.

Main function:

- configure how payroll treats missing attendance;
- configure overtime multipliers;
- configure holiday overtime;
- configure late penalties;
- configure rounding;
- configure paid/unpaid leave rates.

Example body:

```json
{
  "missing_attendance_policy": "PAID_REVIEW_REQUIRED",
  "regular_overtime_multiplier": "1.500000",
  "holiday_overtime_multiplier": "2.000000",
  "late_penalty_mode": "NONE",
  "late_penalty_amount": "0.00",
  "rounding_mode": "CENT",
  "vacation_pay_rate": "1.000000",
  "sick_leave_pay_rate": "1.000000",
  "remote_pay_rate": "1.000000",
  "business_trip_pay_rate": "1.000000",
  "unpaid_leave_pay_rate": "0.000000"
}
```

Important policy options:

| Field | Meaning |
| --- | --- |
| `missing_attendance_policy` | Defines what happens when attendance is missing. |
| `regular_overtime_multiplier` | Multiplier for regular overtime. |
| `holiday_overtime_multiplier` | Multiplier for weekend/holiday overtime. |
| `late_penalty_mode` | Defines late penalty behavior. |
| `rounding_mode` | Defines money rounding. |
| `*_leave_pay_rate` | Defines how different leave types are paid. |

## Salary History

### `POST /payroll/salary-history`

Creates salary history record for an employee.

Main function:

- store salary changes over time;
- calculate salary correctly when employee salary changes mid-month;
- support historical payroll calculations.

Example body:

```json
{
  "employee_id": "EMPLOYEE_ID_HERE",
  "salary_rate": "250000.00",
  "effective_from": "2026-05-01",
  "effective_to": ""
}
```

Important:

Salary history periods must not overlap for the same employee.

## Attendance Overrides

### `PUT /payroll/attendance-overrides`

Creates or updates manual payroll attendance override for one employee and one date.

Main function:

- manually fix payroll attendance for a date;
- mark day as paid or unpaid;
- set attendance status;
- add overtime minutes;
- explain correction with note.

Example body:

```json
{
  "employee_id": "EMPLOYEE_ID_HERE",
  "date": "2026-05-14",
  "paid_day": true,
  "attendance_type": "MANUAL",
  "attendance_status": "PRESENT",
  "overtime_minutes": 90,
  "note": "Manual correction for testing"
}
```

Supported attendance statuses:

```text
PRESENT
LATE
ABSENT
ON_LEAVE
```

## Employment Period

### `PATCH /payroll/employees/{id}/employment-period`

Updates employee hire date and termination date used by payroll.

Main function:

- calculate partial month salary for newly hired employees;
- stop salary calculation after termination date;
- avoid paying outside employment period.

Example body:

```json
{
  "hire_date": "2026-05-01",
  "termination_date": ""
}
```

## Corrections

### `POST /payroll/corrections`

Creates payroll correction for a target cycle.

Main function:

- compensate payroll mistakes from previous cycles;
- add correction bonus;
- add correction deduction;
- connect correction to source payroll item if needed.

Example body:

```json
{
  "employee_id": "EMPLOYEE_ID_HERE",
  "source_item_id": "SOURCE_ITEM_ID_HERE",
  "target_cycle_id": "TARGET_CYCLE_ID_HERE",
  "type": "BONUS",
  "amount": "15000.00",
  "reason": "Correction for previous payroll",
  "taxable": true
}
```

## Review Required Logic

Payroll item can have:

```json
"review_required": true
```

This means HR/Admin must inspect the item before approving the cycle.

Common reasons:

- missing attendance records;
- manual override was used;
- net salary was below zero and clamped to zero;
- unusual payroll data was detected during calculation.

If all items have:

```json
"review_required": false
```

then cycle can be approved directly.

## Full Swagger Testing Example

### 1. Create KZ tax preset

```text
POST /v1/payroll/tax-rules/kz-preset
```

### 2. Create policy

```text
PUT /v1/payroll/policy
```

### 3. Create cycle

```text
POST /v1/payroll/cycles
```

### 4. Preview cycle

```text
POST /v1/payroll/cycles/{cycle_id}/preview
```

### 5. Calculate cycle

```text
PATCH /v1/payroll/cycles/{cycle_id}/calculate
```

### 6. Check items

```text
GET /v1/payroll/cycles/{cycle_id}/items
```

### 7. Review items if needed

```text
PATCH /v1/payroll/items/{item_id}/review
```

### 8. Approve cycle

```text
PATCH /v1/payroll/cycles/{cycle_id}/approve
```

### 9. Prepare payment

```text
PATCH /v1/payroll/cycles/{cycle_id}/prepare-payment
```

### 10. Export payment batch

```text
GET /v1/payroll/cycles/{cycle_id}/export
```

### 11. Mark as paid

```text
PATCH /v1/payroll/cycles/{cycle_id}/mark-paid
```

## What Payroll Does Not Do

The module does not:

- transfer money;
- connect to real bank APIs;
- perform KYC;
- execute legal payments;
- replace accounting approval.

Instead, it prepares payroll data and records payment status after external payment.

This is intentional for an HRMS project and keeps payroll inside the human resource management domain instead of fintech.
