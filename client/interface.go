// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package client

import (
	"context"

	"github.com/harness/gitness/types"
)

// Client to access the remote APIs.
type Client interface {
	// Login authenticates the user and returns a JWT token.
	Login(ctx context.Context, username, password string) (*types.Token, error)

	// Register registers a new  user and returns a JWT token.
	Register(ctx context.Context, username, password string) (*types.Token, error)

	// Self returns the currently authenticated user.
	Self(ctx context.Context) (*types.User, error)

	// Token returns an oauth2 bearer token for the currently
	// authenticated user.
	Token(ctx context.Context) (*types.Token, error)

	// User returns a user by ID or email.
	User(ctx context.Context, key string) (*types.User, error)

	// UserList returns a list of all registered users.
	UserList(ctx context.Context, params types.Params) ([]types.User, error)

	// UserCreate creates a new user account.
	UserCreate(ctx context.Context, user *types.User) (*types.User, error)

	// UserUpdate updates a user account by ID or email.
	UserUpdate(ctx context.Context, key string, input *types.UserInput) (*types.User, error)

	// UserDelete deletes a user account by ID or email.
	UserDelete(ctx context.Context, key string) error
}

// remoteError store the error payload returned
// fro the remote API.
type remoteError struct {
	Message string `json:"message"`
}

// Error returns the error message.
func (e *remoteError) Error() string {
	return e.Message
}