package resolvers

import (
    "context"
    "graphql-pkm/internal/models"
)

func (r *noteResolver) Links(ctx context.Context, obj *models.Note) ([]*models.Link, error) {
    return r.noteService.GetLinks(obj.ID)
}

func (r *noteResolver) Backlinks(ctx context.Context, obj *models.Note) ([]*models.Link, error) {
    return r.noteService.GetBackLinks(obj.ID)
}