-- conversations.sql
-- CRUD operations for conversations table
-- Manages chat sessions with assistants
-- Timestamps managed by DB.

-- name: CreateConversation :execlastid
INSERT INTO conversations (title, assistant_id, conversation_type, linked_note_id, summary, metadata, is_active, last_activity)
VALUES (:title, :assistant_id, :conversation_type, :linked_note_id, :summary, :metadata, :is_active, :last_activity);

-- name: GetConversationByID :one
SELECT * FROM conversations WHERE id = :id;

-- name: ListConversations :many
SELECT * FROM conversations ORDER BY created_at DESC;

-- name: ListActiveConversations :many
SELECT * FROM conversations 
WHERE is_active = 1 
ORDER BY created_at DESC;

-- name: ListConversationsByAssistant :many
SELECT * FROM conversations 
WHERE assistant_id = :assistant_id 
ORDER BY created_at DESC;

-- name: ListConversationsByType :many
SELECT * FROM conversations 
WHERE conversation_type = :conversation_type 
ORDER BY last_activity DESC;

-- name: ListConversationsByLinkedNote :many
SELECT * FROM conversations 
WHERE linked_note_id = :linked_note_id 
ORDER BY last_activity DESC;

-- name: GetRecentConversations :many
SELECT * FROM conversations 
WHERE last_activity >= :since_date
ORDER BY last_activity DESC;

-- name: UpdateConversationByID :exec
UPDATE conversations
SET title = :title,
    assistant_id = :assistant_id,
    conversation_type = :conversation_type,
    linked_note_id = :linked_note_id,
    summary = :summary,
    metadata = :metadata,
    is_active = :is_active,
    last_activity = :last_activity,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: UpdateConversationActivity :exec
UPDATE conversations
SET last_activity = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteConversationByID :exec
DELETE FROM conversations WHERE id = :id;

-- name: SetConversationActive :exec
UPDATE conversations
SET is_active = :is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- ========================================
-- Composite Queries - Conversations with Relations
-- ========================================

-- name: GetConversationWithAssistant :one
SELECT 
    c.*,
    a.id as assistant_id,
    a.name as assistant_name,
    a.provider_type as assistant_provider_type
FROM conversations c
LEFT JOIN assistants a ON c.assistant_id = a.id
WHERE c.id = :id;

-- name: GetConversationWithDetails :one
SELECT 
    c.*,
    a.id as assistant_id,
    a.name as assistant_name,
    a.provider_type as assistant_provider_type,
    a.llm_config as assistant_llm_config,
    COUNT(m.id) as message_count
FROM conversations c
LEFT JOIN assistants a ON c.assistant_id = a.id
LEFT JOIN messages m ON c.id = m.conversation_id
WHERE c.id = :id
GROUP BY c.id;
