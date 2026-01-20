-- name: CreateLinkVisit :one
INSERT INTO link_visits (link_id, created_at, ip, user_agent, referer, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: ListLinkVisitsPage :many
SELECT id, link_id, created_at, ip, user_agent, referer, status
FROM link_visits
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: ListLinkVisits :many
SELECT id, link_id, created_at, ip, user_agent, referer, status
FROM link_visits
ORDER BY created_at DESC, id DESC;

-- name: CountLinkVisits :one
SELECT COUNT(*)
FROM link_visits;
