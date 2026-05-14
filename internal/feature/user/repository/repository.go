package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type UserProfile struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	OrganizationName   string
	Email              string
	Firstname          string
	Lastname           string
	Role               string
	PhoneNumber        string
	VerificationStatus string
	JoinedDate         sql.NullTime
	Department         sql.NullString
	DepartmentID       sql.NullString
	Position           sql.NullString
	Salary             sql.NullString
	Location           sql.NullString
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserProfileByID(ctx context.Context, userID uuid.UUID) (UserProfile, error) {
	const query = `
SELECT
    u.id,
    u.org_id,
    COALESCE(o.name, '') AS organization_name,
    u.email,
    u.first_name,
    u.last_name,
    COALESCE(e.role, u.role) AS role,
    u.phone_number,
    u.verification_status,
    u.created_at AS joined_date,
    d.name AS department_name,
    e.department_id::text AS department_id,
    p.name AS position_name,
    CASE WHEN e.salary_rate IS NOT NULL THEN e.salary_rate::text ELSE NULL END AS salary_rate,
    o.address AS location
FROM users u
LEFT JOIN employees e ON e.user_id = u.id
LEFT JOIN organizations o ON o.id = u.org_id
LEFT JOIN departments d ON d.id = e.department_id
LEFT JOIN positions p ON p.id = e.position_id
WHERE u.id = $1
`

	var profile UserProfile
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.OrganizationID,
		&profile.OrganizationName,
		&profile.Email,
		&profile.Firstname,
		&profile.Lastname,
		&profile.Role,
		&profile.PhoneNumber,
		&profile.VerificationStatus,
		&profile.JoinedDate,
		&profile.Department,
		&profile.DepartmentID,
		&profile.Position,
		&profile.Salary,
		&profile.Location,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserProfile{}, ErrUserNotFound
		}
		return UserProfile{}, err
	}

	return profile, nil
}

func (r *Repository) GetUserRoleByID(ctx context.Context, userID uuid.UUID) (string, error) {
	const query = `
SELECT COALESCE(NULLIF(TRIM(e.role), ''), NULLIF(TRIM(u.role), ''))
FROM users u
LEFT JOIN employees e ON e.user_id = u.id
WHERE u.id = $1
`

	var role sql.NullString
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	if !role.Valid || strings.TrimSpace(role.String) == "" {
		return "", ErrUserNotFound
	}

	return strings.TrimSpace(role.String), nil
}
