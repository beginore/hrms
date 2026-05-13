-- name: InsertOrganization :exec
INSERT INTO organizations (
    id,
    admin_id,
    name,
    vat_id,
    description,
    address,
    city_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: InsertUser :exec
INSERT INTO users (
    id,
    org_id,
    email,
    role,
    first_name,
    last_name,
    phone_number,
    verification_status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'Unverified');

-- name: InsertConsent :exec
INSERT INTO consents (
    id, user_id, org_id, document_type, version,
    created_at
)
VALUES (
           $1, $2, $3, $4, $5, NOW()
       );

-- name: GetUserByEmail :one
SELECT id, org_id, verification_status
FROM users
WHERE email = $1;

-- name: UpdateUserVerificationStatus :exec
UPDATE users
SET verification_status = 'Verified'
WHERE email = $1;

-- name: CheckVATUnique :one
SELECT COUNT(*)
FROM organizations
WHERE vat_id = $1;

-- name: InsertDepartment :exec
INSERT INTO departments (
    id, org_id, name
)
VALUES (
        $1, $2, $3
);

-- name: InsertPosition :exec
INSERT INTO positions (
    id, org_id, name
)
VALUES (
        $1, $2, $3
);

-- name: CountEmployeesByDepartment :one
SELECT COUNT(*)
FROM employees
WHERE department_id = $1;

-- name: CountEmployeesByPosition :one
SELECT COUNT(*)
FROM employees
WHERE position_id = $1;

-- name: GetDepartmentsByOrgID :many
SELECT id, org_id, name FROM departments WHERE org_id = $1 ORDER BY name;

-- name: GetPositionsByOrgID :many
SELECT id, org_id, name FROM positions WHERE org_id = $1 ORDER BY name;

-- name: DeleteDepartment :exec
DELETE FROM departments WHERE id = $1 AND org_id = $2;

-- name: DeletePosition :exec
DELETE FROM positions WHERE id = $1 AND org_id = $2;