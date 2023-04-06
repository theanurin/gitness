// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package repo

import (
	"context"
	"fmt"

	"github.com/harness/gitness/gitrpc"
	apiauth "github.com/harness/gitness/internal/api/auth"
	"github.com/harness/gitness/internal/auth"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"
	"github.com/rs/zerolog/log"
)

// Delete deletes a repo.
func (c *Controller) Delete(ctx context.Context, session *auth.Session, repoRef string) error {
	repo, err := c.repoStore.FindByRef(ctx, repoRef)
	if err != nil {
		return err
	}

	if err = apiauth.CheckRepo(ctx, c.authorizer, session, repo, enum.PermissionRepoDelete, false); err != nil {
		return err
	}

	if err = c.DeleteRepositoryRPC(ctx, session, repo); err != nil {
		return err
	}

	err = c.repoStore.Delete(ctx, repo.ID)
	if err != nil {
		return err
	}
	return nil
}
func (c *Controller) DeleteRepositoryRPC(ctx context.Context, session *auth.Session, repo *types.Repository) error {
	writeParams, err := CreateRPCWriteParams(ctx, c.urlProvider, session, repo)
	if err != nil {
		return fmt.Errorf("failed to create RPC write params: %w", err)
	}

	err = c.gitRPCClient.DeleteRepository(ctx, &gitrpc.DeleteRepositoryParams{
		WriteParams: writeParams,
	})

	// deletion should not fail if dir does not exist in repos dir
	if gitrpc.ErrorStatus(err) == gitrpc.StatusNotFound {
		log.Ctx(ctx).Warn().Msgf("gitrpc repo %s does not exist", repo.GitUID)
	} else if err != nil {
		// deletion has failed before removing(rename) the repo dir
		return fmt.Errorf("gitrpc failed to delete repo %s: %w", repo.GitUID, err)
	}
	return nil
}