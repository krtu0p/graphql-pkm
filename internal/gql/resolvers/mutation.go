package resolvers

import (
    "context"
    "fmt"
    "graphql-pkm/internal/models"
)

func (r *mutationResolver) CreateNote(ctx context.Context, input models.CreateNoteInput) (*models.Note, error) {
    if input.Title == "" {
        return nil, fmt.Errorf("title cannot be empty")
    }
    return r.noteService.CreateNote(input.Title, input.Content, input.Tags)
}

func (r *mutationResolver) UpdateNote(ctx context.Context, id string, input models.UpdateNoteInput) (*models.Note, error) {
    var titlePtr, contentPtr *string
    
    // FIX: Check if pointers are nil, not compare to empty string
    if input.Title != nil {
        titlePtr = input.Title
    }
    if input.Content != nil {
        contentPtr = input.Content
    }
    
    return r.noteService.UpdateNote(id, titlePtr, contentPtr, input.Tags)
}

func (r *mutationResolver) DeleteNote(ctx context.Context, id string) (bool, error) {
    err := r.noteService.DeleteNote(id)
    return err == nil, err
}

// FIX: Change description to *string (pointer)
func (r *mutationResolver) LinkNotes(ctx context.Context, sourceID string, targetID string, description *string) (*models.Link, error) {
    // TODO: Implement linking functionality
    // desc := ""
    // if description != nil {
    //     desc = *description
    // }
    return nil, fmt.Errorf("linking not implemented yet")
}

func (r *mutationResolver) UnlinkNotes(ctx context.Context, linkID string) (bool, error) {
    return false, fmt.Errorf("unlinking not implemented yet")
}