// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package gitrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/harness/gitness/gitrpc/internal/streamio"
	"github.com/harness/gitness/gitrpc/rpc"

	"github.com/rs/zerolog/log"
)

type InfoRefsParams struct {
	// RepoUID is the uid of the git repository
	RepoUID     string
	Service     string
	Options     []string // (key, value) pair
	GitProtocol string
}

func (c *Client) GetInfoRefs(ctx context.Context, w io.Writer, params *InfoRefsParams) error {
	if w == nil {
		return errors.New("writer cannot be nil")
	}
	if params == nil {
		return ErrNoParamsProvided
	}
	stream, err := c.httpService.InfoRefs(ctx, &rpc.InfoRefsRequest{
		RepoUid:          params.RepoUID,
		Service:          params.Service,
		GitConfigOptions: params.Options,
		GitProtocol:      params.GitProtocol,
	})
	if err != nil {
		return fmt.Errorf("error initializing GetInfoRefs() stream: %w", err)
	}

	var (
		response *rpc.InfoRefsResponse
	)
	for {
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("GetInfoRefs() error receiving stream bytes: %w", err)
		}
		_, err = w.Write(response.GetData())
		if err != nil {
			return fmt.Errorf("GetInfoRefs() error: %w", err)
		}
	}

	return nil
}

type ServicePackParams struct {
	// RepoUID is the uid of the git repository
	RepoUID     string
	Service     string
	GitProtocol string
	// PrincipalID used for git hooks in receive-pack service
	PrincipalID int64
	Data        io.ReadCloser
	Options     []string // (key, value) pair
}

func (c *Client) ServicePack(ctx context.Context, w io.Writer, params *ServicePackParams) error {
	if w == nil {
		return errors.New("writer cannot be nil")
	}
	if params == nil {
		return ErrNoParamsProvided
	}

	log := log.Ctx(ctx)

	stream, err := c.httpService.ServicePack(ctx)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Start service pack '%s' with options '%v'.",
		params.Service, params.Options)

	// send basic information
	if err = stream.Send(&rpc.ServicePackRequest{
		RepoUid:          params.RepoUID,
		Service:          params.Service,
		GitConfigOptions: params.Options,
		GitProtocol:      params.GitProtocol,
		PrincipalId:      strconv.FormatInt(params.PrincipalID, 10),
	}); err != nil {
		return err
	}

	log.Debug().Msg("Send request stream.")

	// send body as stream
	stdout := streamio.NewWriter(func(p []byte) error {
		return stream.Send(&rpc.ServicePackRequest{
			Data: p,
		})
	})

	_, err = io.Copy(stdout, params.Data)
	if err != nil {
		return fmt.Errorf("PostUploadPack() error copying reader: %w", err)
	}

	log.Debug().Msg("completed sending request stream.")

	if err = stream.CloseSend(); err != nil {
		return fmt.Errorf("PostUploadPack() error closing the stream: %w", err)
	}

	log.Debug().Msg("start receiving response stream.")

	// when we are done with inputs then we should expect
	// git data
	var (
		response *rpc.ServicePackResponse
	)
	for {
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Debug().Msg("received end of response stream.")
			break
		}
		if err != nil {
			return processRPCErrorf(err, "PostUploadPack() error receiving stream bytes")
		}
		if response.GetData() == nil {
			return fmt.Errorf("PostUploadPack() data is nil")
		}

		_, err = w.Write(response.GetData())
		if err != nil {
			return fmt.Errorf("PostUploadPack() error writing response data: %w", err)
		}
	}

	log.Debug().Msg("completed service pack.")

	return nil
}