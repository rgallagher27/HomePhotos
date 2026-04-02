-- name: CreateTagGroup :one
INSERT INTO tag_groups (name, sort_order)
VALUES (?, ?)
RETURNING id, name, sort_order, created_at;

-- name: GetTagGroupByID :one
SELECT id, name, sort_order, created_at
FROM tag_groups WHERE id = ?;

-- name: UpdateTagGroup :exec
UPDATE tag_groups SET name = ?, sort_order = ? WHERE id = ?;

-- name: DeleteTagGroup :exec
DELETE FROM tag_groups WHERE id = ?;

-- name: ListTagGroups :many
SELECT id, name, sort_order, created_at
FROM tag_groups ORDER BY sort_order, name;

-- name: CreateTag :one
INSERT INTO tags (name, color, group_id, created_by)
VALUES (?, ?, ?, ?)
RETURNING id, name, color, group_id, created_by, created_at;

-- name: GetTagByID :one
SELECT t.id, t.name, t.color, t.group_id, t.created_by, t.created_at,
       tg.name AS group_name
FROM tags t
LEFT JOIN tag_groups tg ON tg.id = t.group_id
WHERE t.id = ?;

-- name: UpdateTag :exec
UPDATE tags SET name = ?, color = ?, group_id = ? WHERE id = ?;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = ?;

-- name: ListTags :many
SELECT t.id, t.name, t.color, t.group_id, t.created_by, t.created_at,
       tg.name AS group_name
FROM tags t
LEFT JOIN tag_groups tg ON tg.id = t.group_id
ORDER BY COALESCE(tg.sort_order, 999999), t.name;

-- name: AddPhotoTag :exec
INSERT OR IGNORE INTO photo_tags (photo_id, tag_id, created_by)
VALUES (?, ?, ?);

-- name: RemovePhotoTag :exec
DELETE FROM photo_tags WHERE photo_id = ? AND tag_id = ?;

-- name: ListTagsForPhoto :many
SELECT t.id, t.name, t.color, t.group_id, t.created_by, t.created_at,
       tg.name AS group_name
FROM photo_tags pt
JOIN tags t ON t.id = pt.tag_id
LEFT JOIN tag_groups tg ON tg.id = t.group_id
WHERE pt.photo_id = ?
ORDER BY t.name;
