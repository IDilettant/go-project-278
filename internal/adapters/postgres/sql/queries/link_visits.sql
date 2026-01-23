-- name: CreateLinkVisit :one
INSERT INTO link_visits (link_id, created_at, ip, user_agent, referer, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: CountLinkVisits :one
SELECT COUNT(*)
FROM link_visits;
