-- assistant_notes.sql
-- CRUD operations for assistant_notes table
-- Assistant's own notes and knowledge base (like user's notes)
-- Follows the same pattern as notes.sql

-- name: CreateAssistantNote :execlastid
INSERT INTO assistant_notes (uuid, title, body, description, note_type, related_conversation_id, related_note_id, created_by_assistant_id, priority, due_date, is_completed, is_active)
VALUES (:uuid, :title, :body, :description, :note_type, :related_conversation_id, :related_note_id, :created_by_assistant_id, :priority, :due_date, :is_completed, :is_active);

-- name: GetAssistantNoteByID :one
SELECT * FROM assistant_notes WHERE id = :id;

-- name: GetAssistantNoteByUUID :one
SELECT * FROM assistant_notes WHERE uuid = :uuid;

-- name: ListAssistantNotes :many
SELECT * FROM assistant_notes ORDER BY created_at DESC;

-- name: ListActiveAssistantNotes :many
SELECT * FROM assistant_notes 
WHERE is_active = 1 
ORDER BY created_at DESC;

-- name: ListAssistantNotesByType :many
SELECT * FROM assistant_notes 
WHERE note_type = :note_type 
ORDER BY created_at DESC;

-- name: ListAssistantNotesByAssistant :many
SELECT * FROM assistant_notes 
WHERE created_by_assistant_id = :created_by_assistant_id 
ORDER BY created_at DESC;

-- name: ListAssistantNotesByConversation :many
SELECT * FROM assistant_notes 
WHERE related_conversation_id = :related_conversation_id 
ORDER BY created_at DESC;

-- name: ListAssistantNotesByLinkedNote :many
SELECT * FROM assistant_notes 
WHERE related_note_id = :related_note_id 
ORDER BY created_at DESC;

-- name: ListAssistantNotesByPriority :many
SELECT * FROM assistant_notes 
WHERE priority >= :min_priority 
ORDER BY priority DESC, created_at DESC;

-- name: ListPendingTasks :many
SELECT * FROM assistant_notes 
WHERE note_type = 'task' 
  AND is_completed = 0 
  AND is_active = 1 
ORDER BY priority DESC, due_date ASC;

-- name: ListUpcomingReminders :many
SELECT * FROM assistant_notes 
WHERE note_type = 'reminder' 
  AND is_completed = 0 
  AND is_active = 1 
  AND due_date <= :before_date 
ORDER BY due_date ASC;

-- name: UpdateAssistantNoteByID :exec
UPDATE assistant_notes
SET uuid = :uuid,
    title = :title,
    body = :body,
    description = :description,
    note_type = :note_type,
    related_conversation_id = :related_conversation_id,
    related_note_id = :related_note_id,
    created_by_assistant_id = :created_by_assistant_id,
    priority = :priority,
    due_date = :due_date,
    is_completed = :is_completed,
    is_active = :is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: MarkAssistantNoteCompleted :exec
UPDATE assistant_notes
SET is_completed = 1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteAssistantNoteByID :exec
DELETE FROM assistant_notes WHERE id = :id;

-- name: SetAssistantNoteActive :exec
UPDATE assistant_notes
SET is_active = :is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: CountAssistantNotes :one
SELECT COUNT(*) FROM assistant_notes;

-- name: DeleteAssistantNotesByLinkedNote :exec
DELETE FROM assistant_notes WHERE related_note_id = :related_note_id;
