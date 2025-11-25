package resolvers

import (
    "graphql-pkm/internal/gql/generated"
    "graphql-pkm/internal/service"
)

type Resolver struct {
    noteService   *service.NoteService
    searchService *service.SearchService
}

func NewResolver(noteService *service.NoteService, searchService *service.SearchService) *Resolver {
    return &Resolver{
        noteService:   noteService,
        searchService: searchService,
    }
}

func (r *Resolver) Query() generated.QueryResolver {
    return &queryResolver{r}
}

func (r *Resolver) Mutation() generated.MutationResolver {
    return &mutationResolver{r}
}

func (r *Resolver) Subscription() generated.SubscriptionResolver {
    return &subscriptionResolver{r}
}

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }