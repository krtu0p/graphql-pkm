package resolvers

import (
    "context"
    "fmt"
    "graphql-pkm/internal/models"
)

func (r *subscriptionResolver) NoteCreated(ctx context.Context) (<-chan *models.Note, error) {
    return nil, fmt.Errorf("subscriptions not implemented yet")
}

func (r *subscriptionResolver) NoteUpdated(ctx context.Context) (<-chan *models.Note, error) {
    return nil, fmt.Errorf("subscriptions not implemented yet")
}