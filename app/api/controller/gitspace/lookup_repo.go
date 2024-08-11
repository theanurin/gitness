// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitspace

import (
	"context"
	"fmt"
	"net/url"

	apiauth "github.com/harness/gitness/app/api/auth"
	"github.com/harness/gitness/app/api/usererror"
	"github.com/harness/gitness/app/auth"
	"github.com/harness/gitness/app/gitspace/scm"
	"github.com/harness/gitness/types/enum"
)

type LookupRepoInput struct {
	Identifier string `json:"-"`
	SpaceRef   string `json:"space_ref"` // Ref of the parent space
	URL        string `json:"url"`
}

var (
	ErrInvalidURL = usererror.BadRequest(
		"The URL specified is not valid format.")
	ErrBadURLScheme = usererror.BadRequest("the URL is missing scheme, it must start with http or https")
)

func (c *Controller) LookupRepo(
	ctx context.Context,
	session *auth.Session,
	in *LookupRepoInput,
) (*scm.CodeRepositoryResponse, error) {
	if err := c.sanitizeLookupRepoInput(in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	space, err := c.spaceStore.FindByRef(ctx, in.SpaceRef)
	if err != nil {
		return nil, fmt.Errorf("failed to find space: %w", err)
	}
	err = apiauth.CheckGitspace(ctx, c.authorizer, session, space.Path, "", enum.PermissionGitspaceEdit)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize: %w", err)
	}
	repositoryRequest := scm.CodeRepositoryRequest{URL: in.URL}
	codeRepositoryResponse, err := c.scm.CheckValidCodeRepo(ctx, repositoryRequest)
	if err != nil {
		return nil, err
	}
	return codeRepositoryResponse, nil
}

func (c *Controller) sanitizeLookupRepoInput(in *LookupRepoInput) error {
	parsedURL, err := url.Parse(in.URL)
	if err != nil {
		return ErrInvalidURL
	}
	if parsedURL.Scheme == "" {
		return ErrBadURLScheme
	}
	if _, err := url.ParseRequestURI(parsedURL.RequestURI()); err != nil {
		return err
	}
	return nil
}
