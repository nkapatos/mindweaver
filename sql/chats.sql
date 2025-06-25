-- name: CreateChat :one
INSERT INTO chats (uuid, conversation_id, actor_id, title, provider_id, model_name, system_prompt_id) 
VALUES (?, ?, ?, ?, ?, ?, ?) 
RETURNING id, uuid, conversation_id, actor_id, title, provider_id, model_name, system_prompt_id, created_at, updated_at;

-- name: GetChatByUUID :one
SELECT id, uuid, conversation_id, actor_id, title, provider_id, model_name, system_prompt_id, created_at, updated_at 
FROM chats 
WHERE uuid = ? 
LIMIT 1;

-- name: GetChatByID :one
SELECT id, uuid, conversation_id, actor_id, title, provider_id, model_name, system_prompt_id, created_at, updated_at 
FROM chats 
WHERE id = ? 
LIMIT 1;

-- name: GetChatsByConversationID :many
SELECT id, uuid, conversation_id, actor_id, title, provider_id, model_name, system_prompt_id, created_at, updated_at 
FROM chats 
WHERE conversation_id = ? 
ORDER BY created_at ASC;

-- name: GetChatsByActorID :many
SELECT id, uuid, conversation_id, actor_id, title, provider_id, model_name, system_prompt_id, created_at, updated_at 
FROM chats 
WHERE actor_id = ? 
ORDER BY created_at DESC;

-- name: UpdateChatTitle :exec
UPDATE chats 
SET title = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: DeleteChat :exec
DELETE FROM chats WHERE id = ?;
