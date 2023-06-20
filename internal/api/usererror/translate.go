// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package usererror

import (
	"errors"
	"net/http"

	"github.com/harness/gitness/gitrpc"
	apiauth "github.com/harness/gitness/internal/api/auth"
	"github.com/harness/gitness/internal/services/webhook"
	"github.com/harness/gitness/store"
	"github.com/harness/gitness/types/check"

	"github.com/harness/go-rbac"
	"github.com/rs/zerolog/log"
)

func Translate(err error) *Error {
	var (
		rError      *Error
		checkError  *check.ValidationError
		gitrpcError *gitrpc.Error
	)

	// TODO: Improve performance of checking multiple errors with errors.Is

	// check if err is RBAC error
	if rbacErr := processRBACErrors(err); rbacErr != nil {
		return rbacErr
	}

	switch {
	// api errors
	case errors.As(err, &rError):
		return rError

	// api auth errors
	case errors.Is(err, apiauth.ErrNotAuthenticated):
		return ErrUnauthorized
	case errors.Is(err, apiauth.ErrNotAuthorized):
		return ErrForbidden

	// validation errors
	case errors.As(err, &checkError):
		return New(http.StatusBadRequest, checkError.Error())

	// store errors
	case errors.Is(err, store.ErrResourceNotFound):
		return ErrNotFound
	case errors.Is(err, store.ErrDuplicate):
		return ErrDuplicate
	case errors.Is(err, store.ErrPrimaryPathCantBeDeleted):
		return ErrPrimaryPathCantBeDeleted
	case errors.Is(err, store.ErrPathTooLong):
		return ErrPathTooLong
	case errors.Is(err, store.ErrNoChangeInRequestedMove):
		return ErrNoChange
	case errors.Is(err, store.ErrIllegalMoveCyclicHierarchy):
		return ErrCyclicHierarchy
	case errors.Is(err, store.ErrSpaceWithChildsCantBeDeleted):
		return ErrSpaceWithChildsCantBeDeleted

	// gitrpc errors
	case errors.As(err, &gitrpcError):
		return NewWithPayload(httpStatusCode(
			gitrpcError.Status),
			gitrpcError.Message,
			gitrpcError.Details,
		)

	// webhook errors
	case errors.Is(err, webhook.ErrWebhookNotRetriggerable):
		return ErrWebhookNotRetriggerable

	// unknown error
	default:
		log.Warn().Msgf("Unable to translate error: %s", err)
		return ErrInternal
	}
}

// lookup of gitrpc error codes to HTTP status codes.
var codes = map[gitrpc.Status]int{
	gitrpc.StatusConflict:           http.StatusConflict,
	gitrpc.StatusInvalidArgument:    http.StatusBadRequest,
	gitrpc.StatusNotFound:           http.StatusNotFound,
	gitrpc.StatusNotImplemented:     http.StatusNotImplemented,
	gitrpc.StatusPreconditionFailed: http.StatusPreconditionFailed,
	gitrpc.StatusUnauthorized:       http.StatusUnauthorized,
	gitrpc.StatusInternal:           http.StatusInternalServerError,
	gitrpc.StatusNotMergeable:       http.StatusPreconditionFailed,
}

// httpStatusCode returns the associated HTTP status code for a gitrpc error code.
func httpStatusCode(code gitrpc.Status) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

func processRBACErrors(err error) *Error {
	msg := err.Error()
	switch {
	case
		errors.Is(err, rbac.ErrBaseURLRequired),
		errors.Is(err, rbac.ErrInvalidPrincipalType),
		errors.Is(err, rbac.ErrAccountRequired),
		errors.Is(err, rbac.ErrPrincipalIdentifierRequired),
		errors.Is(err, rbac.ErrPermissionsRequired),
		errors.Is(err, rbac.ErrResourceTypeRequired),
		errors.Is(err, rbac.ErrResourceTypeKeyRequired),
		errors.Is(err, rbac.ErrResourceTypeValueRequired),
		errors.Is(err, rbac.ErrPermissionRequired),
		errors.Is(err, rbac.ErrPermissionsSizeExceeded),
		errors.Is(err, rbac.ErrInvalidCacheEntryType),
		errors.Is(err, rbac.ErrNoHeader),
		errors.Is(err, rbac.ErrAuthorizationTokenRequired),
		errors.Is(err, rbac.ErrOddNumberOfArguments):
		return New(http.StatusBadRequest, msg)
	case errors.Is(err, rbac.ErrMapperFuncCannotBeNil),
		errors.Is(err, rbac.ErrLoggerCannotBeNil):
		return New(http.StatusInternalServerError, msg)
	}

	return nil
}
