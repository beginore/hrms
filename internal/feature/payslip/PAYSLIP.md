# Payslip Module README

Payslip module is responsible for generating employee payslips from approved payroll results, exporting payslips as downloadable PDF files, sending payslips to employees by email, and keeping payslip history.

The module does not calculate salary by itself. Salary, overtime, bonuses, deductions, taxes, attendance impact, employer taxes, and net salary are calculated by the payroll module first. Payslip takes a finalized payroll item and creates an employee-facing document from it.

## Main Responsibilities

- generate payslip from payroll item;
- store payslip payload snapshot;
- store generated PDF file in the database;
- allow Admin/HR to download PDF;
- allow employee to view and download their own payslip;
- send payslip email with PDF attachment;
- mark sent payslip as `SENT`;
- void incorrect payslip without deleting history;
- regenerate a new payslip after the old one was voided.

## Payslip Source Data

Payslip is generated from:

- payroll cycle;
- payroll item;
- employee profile;
- organization profile;
- payroll adjustments;
- payroll calculation snapshot.

The most important payroll fields copied into payslip are:

- `base_salary`;
- `overtime_amount`;
- `bonuses_total`;
- `deductions_total`;
- `taxes_total`;
- `gross_salary`;
- `net_salary`;
- `employer_taxes_total`;
- `total_employer_cost`;
- attendance summary;
- review data;
- calculation snapshot.

## Payslip Formula

Payslip displays payroll result based on this payroll formula:

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

Payslip does not recalculate this formula. It freezes the result from payroll into `payload_snapshot` and PDF content.

## Status Flow

```text
GENERATED
-> SENT
-> VOID
```

Regeneration creates a new payslip:

```text
VOID old payslip
-> REGENERATE
-> GENERATED new payslip
```

### Status Meaning

| Status | Meaning |
| --- | --- |
| `GENERATED` | Payslip was created and PDF is available for download. |
| `SENT` | Payslip was emailed to employee with PDF attachment. Download is still available. |
| `VOID` | Payslip was cancelled and locked. PDF download and send are blocked. |

## Important Rules

- Payslip can be generated only when payroll cycle is `READY_FOR_PAYMENT` or `PAID`.
- Payroll item must have `review_required=false`.
- One payroll item can have only one active payslip.
- Old voided payslips are kept for history.
- Regenerate works only from a `VOID` payslip.
- `SENT` does not block employee download.
- `VOID` blocks download and send.
- Email sending requires configured SES and verified sender/recipient when SES sandbox is enabled.

## Recommended Testing Flow

```text
1. Create payroll cycle
2. Calculate cycle
3. Review all payroll items with review_required=true
4. Approve cycle
5. Prepare payment
6. Generate payslips
7. Download payslip PDF as Admin/HR
8. Send payslip email
9. Login as employee
10. List my payslips
11. Download my payslip PDF
12. Void payslip if needed
13. Regenerate payslip from voided payslip
```

## Endpoints

All payslip endpoints are protected and require:

```http
Authorization: Bearer <access_token>
```

Base path:

```text
/v1
```

## Admin and HR Endpoints

### `POST /payslips`

Generates one payslip from one payroll item.

Use this endpoint when HR wants to generate a payslip manually for a specific employee payroll item.

Main function:

- creates payslip record;
- creates payload snapshot;
- renders PDF;
- stores PDF in database;
- returns payslip metadata.

Required cycle state:

```text
READY_FOR_PAYMENT or PAID
```

Required payroll item state:

```text
review_required=false
```

Example body:

```json
{
  "payroll_item_id": "PAYROLL_ITEM_ID_HERE"
}
```

Successful response includes:

```json
{
  "id": "PAYSLIP_ID",
  "status": "GENERATED",
  "pdf_filename": "payslip-PAYSLIP_ID.pdf",
  "pdf_generated_at": "2026-05-15T00:00:00Z",
  "pdf_sha256": "..."
}
```

### `POST /payroll/cycles/{id}/payslips/generate`

Generates payslips for all payroll items in a payroll cycle.

Use this endpoint after payroll cycle has been prepared for payment.

Main function:

- loops through cycle payroll items;
- skips items that already have an active payslip;
- skips or blocks items requiring review;
- creates payslips for valid items;
- stores PDF for every created payslip.

Response structure:

```json
{
  "created": [],
  "skipped": []
}
```

`created` means new payslips were generated.

`skipped` means active payslips already existed for those payroll items.

### `GET /payslips`

Lists payslips for the caller's organization.

Main function:

- show payslip history;
- filter payslips by cycle, employee, status, or period;
- find payslip IDs for PDF download, send, void, or regenerate.

Query parameters:

| Parameter | Required | Meaning |
| --- | --- | --- |
| `cycle_id` | no | Filter by payroll cycle. |
| `employee_id` | no | Filter by employee. |
| `status` | no | `GENERATED`, `SENT`, or `VOID`. |
| `period_start` | no | Start date filter, format `YYYY-MM-DD`. |
| `period_end` | no | End date filter, format `YYYY-MM-DD`. |

Example:

```http
GET /v1/payslips?status=GENERATED
```

### `GET /payslips/{id}`

Returns one payslip by ID.

Main function:

- inspect payslip amounts;
- inspect employee and period;
- inspect PDF metadata;
- inspect payload snapshot;
- check sent and void data.

Use it before sending or voiding a payslip.

### `GET /payslips/{id}/pdf`

Downloads payslip PDF as Admin/HR.

Main function:

- returns `application/pdf`;
- sets `Content-Disposition: attachment`;
- allows frontend download button;
- uses stored PDF if it exists;
- renders and stores PDF if old payslip has no stored PDF yet.

Frontend usage:

```text
Admin/HR clicks Download
-> frontend calls GET /v1/payslips/{id}/pdf
-> browser downloads PDF file
```

Blocked when:

```text
status=VOID
```

### `POST /payslips/{id}/send`

Sends payslip to employee email and marks it as `SENT`.

Main function:

- gets employee email from payslip snapshot;
- ensures PDF exists;
- sends email through SES;
- attaches PDF file;
- updates payslip status to `SENT`;
- stores `sent_to_email` and `sent_at`.

Important:

- `SENT` payslip can still be downloaded by Admin/HR and employee;
- if SES is in sandbox, sender and recipient email addresses must be verified;
- email failure does not mean PDF generation failed.

Successful response:

```json
{
  "id": "PAYSLIP_ID",
  "status": "SENT",
  "sent_to_email": "employee@example.com",
  "sent_at": "2026-05-15T00:00:00Z"
}
```

### `POST /payroll/cycles/{id}/payslips/send`

Sends all generated payslips for a payroll cycle.

Main function:

- finds payslips with `status=GENERATED`;
- sends each payslip email with PDF attachment;
- marks successfully sent payslips as `SENT`;
- returns failed items separately.

Response structure:

```json
{
  "sent": [],
  "failed": []
}
```

Use this endpoint when HR wants to send all payslips for a payroll cycle at once.

### `PATCH /payslips/{id}/void`

Voids a payslip without deleting it.

Main function:

- marks payslip as `VOID`;
- stores who voided it;
- stores void timestamp;
- stores void reason;
- keeps payslip for audit/history.

Example body:

```json
{
  "reason": "Incorrect calculation test"
}
```

After void:

- PDF download is blocked;
- send is blocked;
- old payslip remains visible in list;
- regenerate becomes available.

### `POST /payslips/{id}/regenerate`

Creates a new payslip from the same payroll item after the old payslip was voided.

Main function:

- requires old payslip status `VOID`;
- uses the same payroll item;
- generates new payslip ID;
- creates new payload snapshot;
- creates and stores new PDF;
- returns new active payslip.

Use this endpoint when the old payslip was incorrect and HR needs a new version.

If old payslip is not `VOID`, response will be:

```json
{
  "error": "payslip is locked"
}
```

## Employee Endpoints

### `GET /me/payslips`

Lists payslips that belong to the authenticated employee.

Main function:

- employee can see their own payslip history;
- employee cannot see other employees' payslips;
- frontend can use it to render employee payslip page.

Frontend usage:

```text
Employee opens Payslips page
-> frontend calls GET /v1/me/payslips
-> frontend shows list of payslips
```

### `GET /me/payslips/{id}`

Returns one payslip that belongs to the authenticated employee.

Main function:

- show payslip details to employee;
- expose payload snapshot for UI;
- prevent access to other employees' payslips.

### `GET /me/payslips/{id}/pdf`

Downloads authenticated employee payslip PDF.

Main function:

- returns employee's own payslip PDF;
- sets download headers;
- allows frontend employee download button;
- works before and after Admin/HR clicks send;
- blocked only when payslip is `VOID`.

Frontend usage:

```text
Employee clicks Download
-> frontend calls GET /v1/me/payslips/{id}/pdf
-> browser downloads PDF
```

## Swagger Testing Checklist

### 1. Prepare payroll

Create and calculate payroll cycle first.

Required final state before payslip generation:

```text
READY_FOR_PAYMENT
```

If `approve`, `prepare-payment`, or `export` returns:

```json
{
  "error": "payroll cycle is locked"
}
```

Check:

- current payroll cycle status;
- payroll items with `review_required=true`;
- whether you are calling the endpoint in the right order.

### 2. Generate payslips

Call:

```http
POST /v1/payroll/cycles/{cycle_id}/payslips/generate
```

Take payslip ID from:

```json
{
  "created": [
    {
      "id": "PAYSLIP_ID"
    }
  ]
}
```

If payslip already exists, take it from:

```json
{
  "skipped": [
    {
      "id": "PAYSLIP_ID"
    }
  ]
}
```

### 3. Download PDF as Admin/HR

Call:

```http
GET /v1/payslips/{payslip_id}/pdf
```

Expected result:

- Swagger shows `Download file`;
- response type is `application/pdf`;
- file name is similar to `payslip-{id}.pdf`.

### 4. Send email

Call:

```http
POST /v1/payslips/{payslip_id}/send
```

Expected result:

```json
{
  "status": "SENT"
}
```

If SES sandbox blocks email, verify sender and recipient in AWS SES.

### 5. Download as employee

Login as employee and call:

```http
GET /v1/me/payslips
```

Then:

```http
GET /v1/me/payslips/{payslip_id}/pdf
```

Expected result:

- employee downloads their own PDF;
- employee cannot access another employee's payslip.

### 6. Void and regenerate

Void:

```http
PATCH /v1/payslips/{payslip_id}/void
```

Body:

```json
{
  "reason": "Incorrect calculation test"
}
```

Regenerate:

```http
POST /v1/payslips/{payslip_id}/regenerate
```

Expected result:

- old payslip remains `VOID`;
- new payslip is `GENERATED`;
- new payslip has new `id`;
- new PDF metadata is returned.

## Common Errors

### `payroll cycle must be READY_FOR_PAYMENT or PAID`

Payslip generation was called too early.

Fix:

```text
calculate -> review -> approve -> prepare-payment -> generate payslips
```

### `payroll item requires review before payslip generation`

Payroll item has `review_required=true`.

Fix:

```http
PATCH /v1/payroll/items/{item_id}/review
```

### `payslip is locked`

Usually means one of these cases:

- trying to download or send `VOID` payslip;
- trying to regenerate payslip that is not `VOID`;
- trying to void already voided payslip.

### `failed to send payslip`

Usually SES configuration problem.

Check:

- AWS region;
- sender email;
- recipient email;
- SES sandbox verification;
- AWS credentials.

### `payslip not found`

Possible reasons:

- wrong payslip ID;
- caller is from another organization;
- employee tries to access someone else's payslip.

## Frontend Notes

Admin/HR download button:

```http
GET /v1/payslips/{id}/pdf
```

Employee download button:

```http
GET /v1/me/payslips/{id}/pdf
```

Send button:

```http
POST /v1/payslips/{id}/send
```

Void button:

```http
PATCH /v1/payslips/{id}/void
```

Regenerate button:

```http
POST /v1/payslips/{id}/regenerate
```

Frontend should treat PDF download as a binary file response, not JSON.

Recommended frontend behavior:

- show `Download` when status is `GENERATED` or `SENT`;
- hide or disable `Download` when status is `VOID`;
- show `Send` when status is `GENERATED`;
- show `Regenerate` only when status is `VOID`;
- show sent timestamp when status is `SENT`.

## Database Notes

Payslip stores:

- immutable payroll payload snapshot;
- generated PDF bytes;
- PDF filename;
- PDF generation timestamp;
- PDF SHA-256 hash;
- sent email metadata;
- void metadata.

Active payslip uniqueness:

```text
one active payslip per payroll_item_id
```

Voided payslips do not block regeneration.

## Security Notes

- Admin/HR endpoints require payroll/payslip manager role.
- Employee endpoints are scoped to authenticated employee.
- Payslip access is organization-scoped.
- PDF download is blocked for voided payslips.
- Payslip history is preserved for audit.

