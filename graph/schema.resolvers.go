package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/alexanderkarlis/sw-dnsbl/auth"
	"github.com/alexanderkarlis/sw-dnsbl/graph/generated"
	"github.com/alexanderkarlis/sw-dnsbl/graph/model"
	"github.com/alexanderkarlis/sw-dnsbl/middleware"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *mutationResolver) CreateToken(ctx context.Context, data model.UserAuth) (*model.Token, error) {
	t := &model.Token{}

	// Right now , there is only one user for this microservice
	if data.Username == username && data.Password == password {
		token, err := auth.CreateJWT(data.Username, data.Password, 20)
		t.BearerToken = fmt.Sprintf("Bearer %s", token)
		return t, err
	}
	return t, gqlerror.Errorf("invalid credentials")
}

func (r *mutationResolver) Enqueue(ctx context.Context, ips []string) (*bool, error) {
	token := middleware.GetTokenFromContext(ctx)
	log.Println("token:", token)
	if token == "" {
		return nil, gqlerror.Errorf("missing auth token")
	}

	_, err := auth.ValidateToken(strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		return nil, gqlerror.Errorf("not an authorized token")
	}
	isQueueOk := r.Consumer.Queue(ips)

	if !isQueueOk {
		return &isQueueOk, gqlerror.Errorf("queue not started")
	}
	return &isQueueOk, nil
}

func (r *mutationResolver) SetWorkerPoolSize(ctx context.Context, size int) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetIPDetails(ctx context.Context, ip string) (*model.Record, error) {
	token := middleware.GetTokenFromContext(ctx)
	if token == "" {
		return nil, gqlerror.Errorf("missing auth token")
	}

	_, err := auth.ValidateToken(strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		return nil, gqlerror.Errorf("not an authorized token")
	}

	record, err := r.Database.QueryRecord(ip)
	return record, err
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
const (
	username = "secureworks"
	password = "supersecret"
)
