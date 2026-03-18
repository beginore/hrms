# User Module

## Purpose

The `user` module is responsible for returning the current authenticated user's profile for the frontend profile page.

Current endpoint:

- `GET /v1/profile/me`

This endpoint is intended for the profile screen and personal information widgets.

## What the module does

The module:

1. reads the `Authorization` header
2. extracts the access token from `Bearer <token>`
3. parses the Cognito JWT claims to get the current `userId`
4. loads the user profile from PostgreSQL
5. returns a profile response shaped for frontend consumption

## Route

Registered in:

- [main.go](C:/Users/Adel/hrms/cmd/main.go)

Handler:

- [handler.go](C:/Users/Adel/hrms/internal/feature/user/transport/http/handler.go)

Service:

- [service.go](C:/Users/Adel/hrms/internal/feature/user/service/service.go)

Repository:

- [repository.go](C:/Users/Adel/hrms/internal/feature/user/repository/repository.go)

## Request

Method:

- `GET`

URL:

- `/v1/profile/me`

Required header:

```text
Authorization: Bearer <access_token>
```

The endpoint does not accept a request body.

## Response

DTO:

- [dto.go](C:/Users/Adel/hrms/internal/feature/user/service/dto.go)

Returned JSON:

```json
{
  "id": "uuid",
  "organizationId": "uuid",
  "organizationName": "Tech Corp",
  "email": "user@example.com",
  "firstname": "Ivan",
  "lastname": "Ivanov",
  "fullName": "Ivan Ivanov",
  "role": "EMPLOYEE",
  "phone": "+77001234567",
  "phoneNumber": "+77001234567",
  "verificationStatus": "Verified",
  "joinedDate": "2026-03-18T10:00:00Z",
  "department": "Development",
  "position": "Senior Go Developer",
  "salary": "500000.00",
  "location": "Berliner Str 5"
}
```

## Data Sources

The module loads profile data using a SQL query with joins across:

- `users`
- `employees`
- `organizations`
- `departments`
- `positions`

Current query behavior:

- base user data comes from `users`
- organization name and location come from `organizations`
- department, position, salary, and employee role come from `employees` + `departments` + `positions`
- if `employees.role` exists, it overrides `users.role`
- if no employee row exists, profile still works because joins are `LEFT JOIN`

## Field Mapping

Current field sources:

- `id` -> `users.id`
- `organizationId` -> `users.org_id`
- `organizationName` -> `organizations.name`
- `email` -> `users.email`
- `firstname` -> `users.first_name`
- `lastname` -> `users.last_name`
- `fullName` -> computed from `firstname + lastname`
- `role` -> `COALESCE(employees.role, users.role)`
- `phone` -> `users.phone_number`
- `phoneNumber` -> `users.phone_number`
- `verificationStatus` -> `users.verification_status`
- `joinedDate` -> `users.created_at`
- `department` -> `departments.name`
- `position` -> `positions.name`
- `salary` -> `employees.salary_rate`
- `location` -> `organizations.address`

## Current Limitations

This module reflects the current database model, not an ideal future model.

Known limitations:

1. `joinedDate` uses `users.created_at`
   because `employees` does not currently have a `created_at` or `joined_at` column

2. `location` uses `organizations.address`
   because there is no dedicated city/location lookup in the current profile query

3. `department`, `position`, and `salary` are empty if the user has no row in `employees`

4. Salary is returned as text from PostgreSQL
   formatting for UI display should be handled on the frontend

## Error Handling

Possible responses:

- `401 Unauthorized`
  - missing `Authorization` header
  - malformed `Bearer` token
  - invalid or expired access token

- `404 Not Found`
  - user id from token does not exist in `users`

- `500 Internal Server Error`
  - unexpected database or service failure

Examples:

```json
{
  "error": "missing or invalid authorization header"
}
```

```json
{
  "error": "invalid or expired access token"
}
```

```json
{
  "error": "user not found"
}
```

## Frontend Usage

The frontend profile page should call:

- `GET /v1/profile/me`

and use the response instead of deriving profile data from login tokens alone.

Recommended mapping:

```ts
const fullName =
  user?.fullName ||
  [user?.firstname, user?.lastname].filter(Boolean).join(" ") ||
  user?.email ||
  "User";

const userData = {
  email: user?.email || "No email",
  role: user?.role || "EMPLOYEE",
  position: user?.position || "Not specified",
  department: user?.department || "Not specified",
  joinedDate: user?.joinedDate || "Not specified",
  salary: user?.salary || "Not specified",
  phone: user?.phone || user?.phoneNumber || "Not specified",
  location: user?.location || "Not specified",
};
```

## Testing in Postman

1. Login:

`POST /v1/auth/login`

2. Copy `accessToken` from the response

3. Call:

`GET /v1/profile/me`

with header:

```text
Authorization: Bearer <accessToken>
```

## Module Layout

Files:

- [repository.go](C:/Users/Adel/hrms/internal/feature/user/repository/repository.go)
- [dto.go](C:/Users/Adel/hrms/internal/feature/user/service/dto.go)
- [errors.go](C:/Users/Adel/hrms/internal/feature/user/service/errors.go)
- [service.go](C:/Users/Adel/hrms/internal/feature/user/service/service.go)
- [handler.go](C:/Users/Adel/hrms/internal/feature/user/transport/http/handler.go)
