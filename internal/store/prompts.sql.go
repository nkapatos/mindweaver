// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: prompts.sql

package store

import (
	"context"
	"database/sql"
)

const createPrompt = `-- name: CreatePrompt :exec
INSERT INTO prompts (title, content, is_system, created_by, updated_by) VALUES (?, ?, ?, ?, ?)
`

type CreatePromptParams struct {
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	IsSystem  sql.NullInt64 `json:"is_system"`
	CreatedBy int64         `json:"created_by"`
	UpdatedBy int64         `json:"updated_by"`
}

func (q *Queries) CreatePrompt(ctx context.Context, arg CreatePromptParams) error {
	_, err := q.db.ExecContext(ctx, createPrompt,
		arg.Title,
		arg.Content,
		arg.IsSystem,
		arg.CreatedBy,
		arg.UpdatedBy,
	)
	return err
}

const deletePrompt = `-- name: DeletePrompt :exec
DELETE FROM prompts WHERE id = ?
`

func (q *Queries) DeletePrompt(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deletePrompt, id)
	return err
}

const getAllPrompts = `-- name: GetAllPrompts :many
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts
`

func (q *Queries) GetAllPrompts(ctx context.Context) ([]Prompt, error) {
	rows, err := q.db.QueryContext(ctx, getAllPrompts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Prompt
	for rows.Next() {
		var i Prompt
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Content,
			&i.IsSystem,
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

const getPromptById = `-- name: GetPromptById :one
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE id = ? LIMIT 1
`

func (q *Queries) GetPromptById(ctx context.Context, id int64) (Prompt, error) {
	row := q.db.QueryRowContext(ctx, getPromptById, id)
	var i Prompt
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Content,
		&i.IsSystem,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
	)
	return i, err
}

const getPromptsByActorID = `-- name: GetPromptsByActorID :many
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE created_by = ?
`

func (q *Queries) GetPromptsByActorID(ctx context.Context, createdBy int64) ([]Prompt, error) {
	rows, err := q.db.QueryContext(ctx, getPromptsByActorID, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Prompt
	for rows.Next() {
		var i Prompt
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Content,
			&i.IsSystem,
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

const getSystemPrompts = `-- name: GetSystemPrompts :many
SELECT id, title, content, is_system, created_at, updated_at, created_by, updated_by FROM prompts WHERE is_system = 1
`

func (q *Queries) GetSystemPrompts(ctx context.Context) ([]Prompt, error) {
	rows, err := q.db.QueryContext(ctx, getSystemPrompts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Prompt
	for rows.Next() {
		var i Prompt
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Content,
			&i.IsSystem,
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

const updatePrompt = `-- name: UpdatePrompt :exec
UPDATE prompts SET title = ?, content = ?, is_system = ?, updated_at = CURRENT_TIMESTAMP, updated_by = ? WHERE id = ?
`

type UpdatePromptParams struct {
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	IsSystem  sql.NullInt64 `json:"is_system"`
	UpdatedBy int64         `json:"updated_by"`
	ID        int64         `json:"id"`
}

func (q *Queries) UpdatePrompt(ctx context.Context, arg UpdatePromptParams) error {
	_, err := q.db.ExecContext(ctx, updatePrompt,
		arg.Title,
		arg.Content,
		arg.IsSystem,
		arg.UpdatedBy,
		arg.ID,
	)
	return err
}
