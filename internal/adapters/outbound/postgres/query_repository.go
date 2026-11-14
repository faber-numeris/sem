package postgres

import (
	"context"

	"github.com/faber-numeris/sem/internal/adapters/outbound/postgres/gen"
	"github.com/faber-numeris/sem/internal/core/domain"
)

func mapCursor(c gen.SalesforceQueryCursor) *domain.QueryCursor {
	return &domain.QueryCursor{
		ID:             c.ID,
		SObjectName:    c.SobjectName,
		SOQLQuery:      c.SoqlQuery,
		IncludeDeleted: c.IncludeDeleted,
		OffsetValue:    c.OffsetValue,
		BatchSize:      c.BatchSize,
		TotalSize:      c.TotalSize,
	}
}

func (r *Repository) CreateQueryCursor(ctx context.Context, cursor *domain.QueryCursor) (*domain.QueryCursor, error) {
	created, err := r.queries.CreateQueryCursor(ctx, gen.CreateQueryCursorParams{
		ID:             cursor.ID,
		SobjectName:    cursor.SObjectName,
		SoqlQuery:      cursor.SOQLQuery,
		IncludeDeleted: cursor.IncludeDeleted,
		OffsetValue:    cursor.OffsetValue,
		BatchSize:      cursor.BatchSize,
		TotalSize:      cursor.TotalSize,
	})
	if err != nil {
		return nil, err
	}
	return mapCursor(created), nil
}

func (r *Repository) GetQueryCursor(ctx context.Context, id string) (*domain.QueryCursor, error) {
	c, err := r.queries.GetQueryCursor(ctx, id)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return mapCursor(c), nil
}

func (r *Repository) UpdateQueryCursorOffset(ctx context.Context, id string, offset int) (*domain.QueryCursor, error) {
	c, err := r.queries.UpdateQueryCursorOffset(ctx, gen.UpdateQueryCursorOffsetParams{
		OffsetValue: offset,
		ID:          id,
	})
	if err != nil {
		return nil, err
	}
	return mapCursor(c), nil
}

func (r *Repository) DeleteQueryCursor(ctx context.Context, id string) error {
	return r.queries.DeleteQueryCursor(ctx, id)
}
