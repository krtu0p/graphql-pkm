package resolvers

import (
    "context"
    "fmt"
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
	links, err := r.noteService.GetBackLinks(noteID)
	if err != nil {
		return nil, err
	}

	backlinks := make([]*models.Backlink, len(links))
	for i, link := range links {
		sourceNote, err := r.noteService.GetNote(link.SourceNoteID)
		if err != nil {
			return nil, fmt.Errorf("failed to get source note: %w", err)
		}

		backlinks[i] = &models.Backlink{
			ID:          link.ID,
			SourceNote:  sourceNote,
			Description: link.Description,
			CreatedAt:   link.CreatedAt,
		}
	}

	return backlinks, nil
}

func (r *queryResolver) Links(ctx context.Context, noteID string) ([]*models.Link, error) {
    return r.noteService.GetLinks(noteID)
}

func (r *queryResolver) CheckRelationship(ctx context.Context, noteIDA string, noteIDB string) (string, error) {
    return r.searchService.CheckRelationship(noteIDA, noteIDB)
}