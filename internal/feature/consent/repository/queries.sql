-- name: GetActiveDocuments :many
SELECT type, version, url
FROM documents
WHERE is_active = true;

-- name: InsertConsentForOrg :exec
INSERT INTO consents (id, user_id, org_id, document_type, version, created_at)
VALUES ($1, $2, $3, $4, $5, NOW());

-- name: GetConsentsByOrgID :many
SELECT DISTINCT ON (document_type) document_type, version
FROM consents
WHERE org_id = $1
ORDER BY document_type, created_at DESC;

-- name: GetLatestDocumentVersions :many
SELECT type, version
FROM documents
WHERE is_active = true;

-- name: GetAdminIDByOrgID :one
SELECT admin_id FROM organizations WHERE id = $1;