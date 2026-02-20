package usecase

import (
	"context"
	"fmt"

	"github.com/turahe/pkg/domain"
	"github.com/turahe/pkg/domain/port"
)

// GetItemByID runs the "get item by ID" use case. It depends only on the domain port.
// Returned error is suitable for handler.HandleServiceError (e.g. domain.ErrNotFound).
func GetItemByID(ctx context.Context, repo port.GetByID, id string) (interface{}, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository: %w", domain.ErrNotFound)
	}
	entity, found, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("id %s: %w", id, domain.ErrNotFound)
	}
	return entity, nil
}
