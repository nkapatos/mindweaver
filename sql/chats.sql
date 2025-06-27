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

-- name: GetChatWithProviderDetails :one
SELECT c.id, c.uuid, c.conversation_id, c.actor_id, c.title, c.provider_id, c.model_name, c.system_prompt_id, c.created_at, c.updated_at,
       p.name as provider_name, p.description as provider_description, p.system_prompt as provider_system_prompt,
       ls.name as llm_service_name, ls.adapter, ls.base_url, ls.organization, ls.configuration
FROM chats c
LEFT JOIN providers p ON c.provider_id = p.id
LEFT JOIN llm_services ls ON p.llm_service_id = ls.id
WHERE c.id = ?
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
