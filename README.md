# Local setup

1. Install the migration tool:

   `go install github.com/pressly/goose/v3/cmd/goose@latest`
2. Apply migration:

   `goose -dir internal/infrastructure/storage/postgres/migrations postgres "postgres://postgres:postgres@localhost:5432/iam?sslmode=disable" up` or `make migrate-up`
3. Run the application:
4.
`go run .\cmd\main.go ".\configs\local\config.toml"`

## Invite SMTP

Invites use SMTP configuration from the `[smtp]` section.

For Gmail SMTP use:

`host = "smtp.gmail.com"`

`port = 587`

`username = "your-email@gmail.com"`

`password = "your-gmail-app-password"`

`sender_email = "your-email@gmail.com"`

Gmail requires an App Password when 2-Step Verification is enabled.

## Invite Flow

The invite system is used for employee activation. Employees cannot self-register. They can join only through an invitation created by an administrator.

### Prerequisites

Before testing invites, make sure that:

1. PostgreSQL is running
2. migrations are applied, including the `invites` table migration
3. the application is running on `http://localhost:8080`
4. SMTP is configured in `configs/local/config.toml`
5. Cognito configuration points to a real user pool and app client
6. there is at least one record in the `organizations` table

Get organization ids with:

```sql
SELECT id, name
FROM organizations;
```

### High-Level Flow

1. Admin creates an invite with employee information
2. The backend generates a unique one-time code
3. The invite is stored in the `invites` table with a 24 hour expiration time
4. The backend sends the invite email through SMTP
5. The employee enters the code on the frontend
6. The frontend calls `POST /v1/invites/verify`
7. If the code is valid, the frontend shows the activation screen with employee and company information
8. The employee enters a password and phone number
9. The frontend calls `POST /v1/invites/complete-registration`
10. The backend creates the user in Cognito, creates the application user in PostgreSQL, and marks the invite as used

### Invite Endpoints

#### 1. Generate Invite

Endpoint:

`POST /v1/invites/generate`

Request body:

```json
{
  "organizationId": "PUT-ORG-UUID-HERE",
  "firstName": "Adel",
  "lastName": "Kenesova",
  "email": "employee@example.com",
  "role": "Employee",
  "position": "HR Manager"
}
```

Successful response:

```json
{
  "inviteId": "uuid",
  "organizationId": "uuid",
  "organizationName": "Schmidt GmbH",
  "email": "employee@example.com",
  "code": "ABCD-1234",
  "expiresAt": "2026-03-18T10:00:00Z"
}
```

What happens:

1. request is validated
2. organization is checked
3. invite code is generated
4. invite is saved in PostgreSQL
5. invite email is sent through SMTP

Important:

- if SMTP send fails, the invite record is deleted
- invite code format is `ABCD-1234`
- invite code is one-time use

#### 2. Verify Invite

Endpoint:

`POST /v1/invites/verify`

Request body:

```json
{
  "code": "ABCD-1234"
}
```

Successful response:

```json
{
  "organizationId": "uuid",
  "organizationName": "Schmidt GmbH",
  "firstName": "Adel",
  "lastName": "Kenesova",
  "fullName": "Adel Kenesova",
  "email": "employee@example.com",
  "role": "Employee",
  "position": "HR Manager",
  "expiresAt": "2026-03-18T10:00:00Z",
  "message": "You have been invited to join Schmidt GmbH"
}
```

This response is intended for the activation page UI. The frontend can display:

- organization name
- employee full name
- email
- position

The backend checks:

- code exists
- code is not expired
- code was not already used

#### 3. Complete Registration

Endpoint:

`POST /v1/invites/complete-registration`

Request body:

```json
{
  "code": "ABCD-1234",
  "password": "StrongPass123!",
  "phoneNumber": "+77001234567"
}
```

Successful response:

```json
{
  "userId": "uuid",
  "organizationId": "uuid",
  "role": "Employee"
}
```

What happens:

1. invite is loaded and locked in a database transaction
2. invite validity is checked again
3. Cognito user is created with `AdminCreateUser`
4. password is set with `AdminSetUserPassword`
5. application user is inserted into the `users` table with `verification_status = 'Verified'`
6. invite is marked as used

### Activation Page Contract

The intended frontend flow is:

1. user enters invite code
2. frontend calls `POST /v1/invites/verify`
3. frontend renders the activation page using the response data
4. user enters only:
   - password
   - phone number
5. frontend calls `POST /v1/invites/complete-registration`

The backend already takes first name, last name, email, role, and position from the invite itself.

### Validation Rules

Password:

- minimum 8 characters in backend pre-validation
- must contain at least one special character
- Cognito may enforce additional password policy rules such as requiring numeric characters

Phone number:

- required for complete registration
- must be in international format
- expected format example: `+77001234567`

Invite:

- one-time use
- expires after 24 hours

### Error Handling

Common responses:

- `400 Bad Request`
  - invalid request body
  - invalid email
  - invalid phone number
  - password does not match policy
- `404 Not Found`
  - organization not found
  - invite not found
- `409 Conflict`
  - invite already used
  - invite expired
  - user already exists
  - phone number already exists
  - email already exists
- `503 Service Unavailable`
  - SMTP is not configured

Examples:

```json
{
  "error": "phone number must be in international format, for example +77001234567"
}
```

```json
{
  "error": "Password does not conform to policy: Password must have numeric characters"
}
```

```json
{
  "error": "phone number already exists"
}
```

### Recovery and Rollback Behavior

The invite flow includes rollback behavior for Cognito in failure scenarios:

- if password setup fails after Cognito user creation, the created Cognito user is deleted
- if PostgreSQL user creation fails after Cognito user creation, the created Cognito user is deleted

This helps prevent partially created employees in Cognito.

### Useful Database Checks

Latest invites:

```sql
SELECT id, email, code, is_used, expires_at, used_at
FROM invites
ORDER BY created_at DESC;
```

Users:

```sql
SELECT id, email, phone_number, verification_status
FROM users
ORDER BY created_at DESC;
```

### Logs

The invite module writes detailed logs for:

- invite generation
- invite verification
- SMTP send attempts
- Cognito user creation
- password setup
- PostgreSQL user insert
- invite usage update

Useful log prefixes:

- `[Invite Generate]`
- `[Invite Verify]`
- `[Invite CompleteRegistration]`
- `[Invite Cognito]`
- `[Invite Mailer]`
