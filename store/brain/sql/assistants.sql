-- assistants.sql
-- CRUD operations for assistants table
-- Manages named AI assistants that users create
-- Examples: "My Code Buddy", "Research Assistant", "Writing Helper"

-- name: CreateAssistant :execlastid
INSERT INTO assistants (name, description, provider_type, api_key, base_url, organization, llm_config, system_prompt_id, is_active)
VALUES (:name, :description, :provider_type, :api_key, :base_url, :organization, :llm_config, :system_prompt_id, :is_active);

-- name: GetAssistantByID :one
SELECT * FROM assistants WHERE id = :id;

-- name: GetAssistantByName :one
SELECT * FROM assistants WHERE name = :name LIMIT 1;

-- name: ListAssistants :many
SELECT * FROM assistants ORDER BY created_at DESC;

-- name: ListActiveAssistants :many
SELECT * FROM assistants 
WHERE is_active = 1 
ORDER BY created_at DESC;

-- name: ListAssistantsByProviderType :many
SELECT * FROM assistants 
WHERE provider_type = :provider_type 
ORDER BY created_at DESC;

-- name: UpdateAssistantByID :exec
UPDATE assistants
SET name = :name,
    description = :description,
    provider_type = :provider_type,
    api_key = :api_key,
    base_url = :base_url,
    organization = :organization,
    llm_config = :llm_config,
    system_prompt_id = :system_prompt_id,
    is_active = :is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- name: DeleteAssistantByID :exec
DELETE FROM assistants WHERE id = :id;

-- name: SetAssistantActive :exec
UPDATE assistants
SET is_active = :is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = :id;

-- ========================================
-- Composite Queries - Assistants with Relations
-- ========================================

-- name: GetAssistantWithPrompt :one
SELECT 
    a.*,
    p.id as prompt_id,
    p.title as prompt_title,
    p.content as prompt_content,
    p.category as prompt_category
FROM assistants a
LEFT JOIN prompts p ON a.system_prompt_id = p.id
WHERE a.id = :id;
