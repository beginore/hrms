-- name: InsertEmployee :exec
INSERT INTO employees (
    id,
    org_id,
    user_id,
    department_id,
    position_id,
    role,
    salary_rate,
    status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetEmployeeByID :one
SELECT
    e.id,
    e.org_id,
    e.user_id,
    e.department_id,
    e.position_id,
    e.role,
    e.salary_rate,
    e.status,
    u.first_name,
    u.last_name,
    u.email,
    u.phone_number,
    d.name AS department_name,
    p.name AS position_name
FROM employees e
         JOIN users       u ON u.id = e.user_id
         JOIN departments d ON d.id = e.department_id
         JOIN positions   p ON p.id = e.position_id
WHERE e.id = $1;

-- name: GetEmployeesByOrgID :many
SELECT
    e.id,
    e.org_id,
    e.user_id,
    e.department_id,
    e.position_id,
    e.role,
    e.salary_rate,
    e.status,
    u.first_name,
    u.last_name,
    u.email,
    u.phone_number,
    d.name AS department_name,
    p.name AS position_name
FROM employees e
         JOIN users       u ON u.id = e.user_id
         JOIN departments d ON d.id = e.department_id
         JOIN positions   p ON p.id = e.position_id
WHERE e.org_id = $1
ORDER BY u.last_name, u.first_name;

-- name: UpdateEmployeeRole :exec
UPDATE employees
SET role = $2
WHERE id = $1;

-- name: UpdateEmployeeSalary :exec
UPDATE employees
SET salary_rate = $2
WHERE id = $1;

-- name: UpdateEmployeeStatus :exec
UPDATE employees
SET status = $2
WHERE id = $1;

-- name: UpdateEmployeeDepartment :exec
UPDATE employees
SET department_id = $2
WHERE id = $1;

-- name: UpdateEmployeePosition :exec
UPDATE employees
SET position_id = $2
WHERE id = $1;

-- name: DeleteEmployee :exec
DELETE FROM employees
WHERE id = $1;

-- name: CheckEmployeeExists :one
SELECT COUNT(*)
FROM employees
WHERE id = $1;

-- name: GetOrgIDByUserID :one
SELECT org_id FROM users WHERE id = $1;
