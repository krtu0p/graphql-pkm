package resolvers

import (
    "context"
    "graphql-pkm/internal/models"
)

func (r *mutationResolver) CreateNote(ctx context.Context, input models.CreateNoteInput) (*models.Note, error) {
    return r.noteService.CreateNote(input.Title, input.Content, input.Tags)
}

func (r *mutationResolver) UpdateNote(ctx context.Context, id string, input models.UpdateNoteInput) (*models.Note, error) {
    return r.noteService.UpdateNote(id, input.Title, input.Content, input.Tags)
}

func (r *mutationResolver) DeleteNote(ctx context.Context, id string) (bool, error) {
    err := r.noteService.DeleteNote(id)
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *mutationResolver) LinkNotes(ctx context.Context, sourceID string, targetID string, description *string) (*models.Link, error) {
    desc := ""
    if description != nil {
        desc = *description
    }
    
    return r.noteService.CreateLink(sourceID, targetID, desc)
}

func (r *mutationResolver) UnlinkNotes(ctx context.Context, linkID string) (bool, error) {
    err := r.noteService.DeleteLink(linkID)
    if err != nil {
        return false, err
    }
    return true, nil
}