package resolvers

import (
    "context"
    "graphql-pkm/internal/models"
)

func (r *queryResolver) Notes(ctx context.Context) ([]*models.Note, error) {
    return r.noteService.GetAllNotes()
}

func (r *queryResolver) Note(ctx context.Context, id string) (*models.Note, error) {
    return r.noteService.GetNote(id)
}

func (r *queryResolver) SearchNotes(ctx context.Context, query string) ([]*models.Note, error) {
    if query == "" {
        return []*models.Note{}, nil
    }
    return r.searchService.SearchNotes(query)
}

func (r *queryResolver) NotesByTag(ctx context.Context, tag string) ([]*models.Note, error) {
    if tag == "" {
        return []*models.Note{}, nil
    }
    return r.noteService.GetNotesByTag(tag)
}

func (r *queryResolver) Backlinks(ctx context.Context, noteID string) ([]*models.Backlink, error) {
    return []*models.Backlink{}, nil
}

func (r *queryResolver) Links(ctx context.Context, noteID string) ([]*models.Link, error) {
    return r.noteService.GetLinks(noteID)
}