package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateInviteParams struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Code      string
	Role      string
	Position  *string
	ExpiresAt time.Time
}

type CreateInvitedUserParams struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	Email       string
	Role        string
	FirstName   string
	LastName    string
	PhoneNumber string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateInvite(ctx context.Context, params CreateInviteParams) error {
	const query = `
INSERT INTO invites (
    id,
    org_id,
    first_name,
    last_name,
    email,
    code,
    role,
    position,
    expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

	_, err := r.db.ExecContext(
		ctx,
		query,
		params.ID,
		params.OrgID,
		params.FirstName,
		params.LastName,
		params.Email,
		params.Code,
		params.Role,
		params.Position,
		params.ExpiresAt,
	)

	return err
}

func (r *Repository) DeleteInviteByID(ctx context.Context, inviteID uuid.UUID) error {
	const query = `DELETE FROM invites WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, inviteID)
	return err
}

func (r *Repository) GetOrganizationNameByID(ctx context.Context, orgID uuid.UUID) (string, error) {
	const query = `SELECT name FROM organizations WHERE id = $1`

	var organizationName string
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&organizationName)
	return organizationName, err
}

func (r *Repository) GetInviteByCode(ctx context.Context, code string) (Invite, error) {
	const query = `
SELECT i.id, i.org_id, o.name, i.first_name, i.last_name, i.email, i.code, i.role, i.position, i.expires_at, i.is_used, i.used_at, i.created_at
FROM invites i
JOIN organizations o ON o.id = i.org_id
WHERE i.code = $1
`

	return scanInvite(r.db.QueryRowContext(ctx, query, code))
}

func (r *Repository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *Repository) GetInviteByCodeTx(ctx context.Context, tx *sql.Tx, code string) (Invite, error) {
	const query = `
SELECT i.id, i.org_id, o.name, i.first_name, i.last_name, i.email, i.code, i.role, i.position, i.expires_at, i.is_used, i.used_at, i.created_at
FROM invites i
JOIN organizations o ON o.id = i.org_id
WHERE i.code = $1
FOR UPDATE
`

	return scanInvite(tx.QueryRowContext(ctx, query, code))
}

func (r *Repository) InsertUserTx(ctx context.Context, tx *sql.Tx, params CreateInvitedUserParams) error {
	const query = `
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
VALUES ($1, $2, $3, $4, $5, $6, $7, 'Verified')
`

	_, err := tx.ExecContext(
		ctx,
		query,
		params.ID,
		params.OrgID,
		params.Email,
		params.Role,
		params.FirstName,
		params.LastName,
		params.PhoneNumber,
	)

	return err
}

func (r *Repository) MarkInviteUsedTx(ctx context.Context, tx *sql.Tx, inviteID uuid.UUID, usedAt time.Time) error {
	const query = `
UPDATE invites
SET is_used = true, used_at = $2
WHERE id = $1
`

	_, err := tx.ExecContext(ctx, query, inviteID, usedAt)
	return err
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return err != nil && errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func scanInvite(scanner interface {
	Scan(dest ...any) error
}) (Invite, error) {
	var invite Invite
	var position sql.NullString

	err := scanner.Scan(
		&invite.ID,
		&invite.OrgID,
		&invite.OrganizationName,
		&invite.FirstName,
		&invite.LastName,
		&invite.Email,
		&invite.Code,
		&invite.Role,
		&position,
		&invite.ExpiresAt,
		&invite.IsUsed,
		&invite.UsedAt,
		&invite.CreatedAt,
	)
	if err != nil {
		return Invite{}, err
	}

	if position.Valid {
		invite.Position = &position.String
	}

	return invite, nil
}
