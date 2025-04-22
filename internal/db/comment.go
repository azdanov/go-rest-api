package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/azdanov/go-rest-api/internal/comment"
)

type CommentRow struct {
	ID     string
	Slug   sql.NullString
	Body   sql.NullString
	Author sql.NullString
}

func convertRowToComment(cr CommentRow) comment.Comment {
	return comment.Comment{
		ID:     cr.ID,
		Slug:   cr.Slug.String,
		Body:   cr.Body.String,
		Author: cr.Author.String,
	}
}

func convertCommentToRow(c comment.Comment) CommentRow {
	return CommentRow{
		ID:     c.ID,
		Slug:   sql.NullString{String: c.Slug, Valid: true},
		Body:   sql.NullString{String: c.Body, Valid: true},
		Author: sql.NullString{String: c.Author, Valid: true},
	}
}

func (d *Database) GetComment(ctx context.Context, id string) (comment.Comment, error) {
	var cr CommentRow
	row := d.Client.QueryRowContext(
		ctx,
		"SELECT id, slug, body, author FROM comments WHERE id = $1",
		id,
	)
	err := row.Scan(&cr.ID, &cr.Slug, &cr.Body, &cr.Author)
	if err != nil {
		return comment.Comment{}, fmt.Errorf("failed to scan comment row: %w", err)
	}

	return convertRowToComment(cr), nil
}

func (d *Database) CreateComment(ctx context.Context, c comment.Comment) (comment.Comment, error) {
	cr := convertCommentToRow(c)

	rows, err := d.Client.NamedQueryContext(
		ctx,
		"INSERT INTO comments (id, slug, body, author) VALUES (:id, :slug, :body, :author)",
		cr,
	)
	if err != nil {
		return comment.Comment{}, fmt.Errorf("failed to insert comment: %w", err)
	}
	if rows.Err() != nil {
		return comment.Comment{}, fmt.Errorf("failed to execute insert query: %w", rows.Err())
	}
	defer rows.Close()

	return c, nil
}

func (d *Database) UpdateComment(ctx context.Context, c comment.Comment) error {
	cr := convertCommentToRow(c)

	rows, err := d.Client.NamedQueryContext(
		ctx,
		"UPDATE comments SET slug = :slug, body = :body, author = :author WHERE id = :id",
		cr,
	)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("failed to execute update query: %w", rows.Err())
	}
	defer rows.Close()

	return nil
}

func (d *Database) DeleteComment(ctx context.Context, id string) error {
	_, err := d.Client.ExecContext(
		ctx,
		"DELETE FROM comments WHERE id = $1",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}
