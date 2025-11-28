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
	return r.Resolver.noteService.CreateNote(input.Title, input.Content, input.Tags)  // r.Resolver
}

func (r *mutationResolver) UpdateNote(ctx context.Context, id string, input models.UpdateNoteInput) (*models.Note, error) {
    var titlePtr, contentPtr *string
    
    // FIX: Check if pinters are nil, not compare to empty string
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

func (r *mutationResolver) LinkNotes(ctx context.Context, sourceID string, targetID string, description *string) (*models.Link, error) {
    desc := ""
    if description != nil {
        desc = *description
    }
    
    link, err := r.noteService.CreateLink(sourceID, targetID, desc)
    if err != nil {
        return nil, err
    }
    
    return link, nil
}

func (r *mutationResolver) UnlinkNotes(ctx context.Context, linkID string) (bool, error) {
    err := r.noteService.DeleteLink(linkID)
    if err != nil {
        return false, err
    }
    return true, nil
}