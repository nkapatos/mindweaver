// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: conversations.sql

package store

import (
	"context"
	"database/sql"
)

const createConversation = `-- name: CreateConversation :one
INSERT INTO conversations (title, description, is_active, metadata, created_by, updated_by) 
VALUES (?, ?, ?, ?, ?, ?) 
RETURNING id, title, description, is_active, metadata, created_at, updated_at, created_by, updated_by
`

type CreateConversationParams struct {
	Title       string         `json:"title"`
	Description sql.NullString `json:"description"`
	IsActive    sql.NullBool   `json:"is_active"`
	Metadata    sql.NullString `json:"metadata"`
	CreatedBy   int64          `json:"created_by"`
	UpdatedBy   int64          `json:"updated_by"`
}

func (q *Queries) CreateConversation(ctx context.Context, arg CreateConversationParams) (Conversation, error) {
	row := q.db.QueryRowContext(ctx, createConversation,
		arg.Title,
		arg.Description,
		arg.IsActive,
		arg.Metadata,
		arg.CreatedBy,
		arg.UpdatedBy,
	)
	var i Conversation
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.IsActive,
		&i.Metadata,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
	)
	return i, err
}

const deleteConversation = `-- name: DeleteConversation :exec
DELETE FROM conversations WHERE id = ?
`

func (q *Queries) DeleteConversation(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteConversation, id)
	return err
}

const getConversationByID = `-- name: GetConversationByID :one
SELECT id, title, description, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM conversations 
WHERE id = ? 
LIMIT 1
`

func (q *Queries) GetConversationByID(ctx context.Context, id int64) (Conversation, error) {
	row := q.db.QueryRowContext(ctx, getConversationByID, id)
	var i Conversation
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.IsActive,
		&i.Metadata,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
	)
	return i, err
}

const getConversationsByActorID = `-- name: GetConversationsByActorID :many
SELECT id, title, description, is_active, metadata, created_at, updated_at, created_by, updated_by 
FROM conversations 
WHERE created_by = ? AND is_active = true 
ORDER BY created_at DESC
`

func (q *Queries) GetConversationsByActorID(ctx context.Context, createdBy int64) ([]Conversation, error) {
	rows, err := q.db.QueryContext(ctx, getConversationsByActorID, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Conversation
	for rows.Next() {
		var i Conversation
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.IsActive,
			&i.Metadata,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CreatedBy,
			&i.UpdatedBy,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateConversation = `-- name: UpdateConversation :exec
UPDATE conversations 
SET title = ?, description = ?, is_active = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? 
WHERE id = ?
`

type UpdateConversationParams struct {
	Title       string         `json:"title"`
	Description sql.NullString `json:"description"`
	IsActive    sql.NullBool   `json:"is_active"`
	Metadata    sql.NullString `json:"metadata"`
	UpdatedBy   int64          `json:"updated_by"`
	ID          int64          `json:"id"`
}

func (q *Queries) UpdateConversation(ctx context.Context, arg UpdateConversationParams) error {
	_, err := q.db.ExecContext(ctx, updateConversation,
		arg.Title,
		arg.Description,
		arg.IsActive,
		arg.Metadata,
		arg.UpdatedBy,
		arg.ID,
	)
	return err
}
