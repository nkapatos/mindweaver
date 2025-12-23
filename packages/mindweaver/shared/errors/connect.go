// Connect-RPC Error Helpers
//
// This file provides reusable error builders for Connect-RPC handlers that follow
// Google's AIP-193 error response standard with rich error details.
//
// Reference: https://google.aip.dev/193
package errors

import (
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// Domain constants for ErrorInfo per Google AIP-193 standards.
// These identify which service domain the error originates from.
const (
	MindDomain  = "mind.mindweaver.com"
	BrainDomain = "brain.mindweaver.com"
)

// NewNotFoundError creates a NOT_FOUND error with ErrorInfo details.
//
// Usage in handlers:
//
//	if errors.Is(err, ErrNoteNotFound) {
//	    return nil, errors.NewNotFoundError(errors.MindDomain, "note", noteID)
//	}
func NewNotFoundError(domain, resource, resourceID string) error {
	err := connect.NewError(
		connect.CodeNotFound,
		fmt.Errorf("%s not found: %s", resource, resourceID),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "RESOURCE_NOT_FOUND",
		Domain: domain,
		Metadata: map[string]string{
			"resource": resource,
			"id":       resourceID,
		},
	})
	err.AddDetail(detail)

	return err
}

// NewAlreadyExistsError creates an ALREADY_EXISTS error with ErrorInfo details.
//
// Usage in handlers:
//
//	if errors.Is(err, ErrNoteAlreadyExists) {
//	    return nil, errors.NewAlreadyExistsError(errors.MindDomain, "note", "title", noteTitle)
//	}
func NewAlreadyExistsError(domain, resource, field, value string) error {
	err := connect.NewError(
		connect.CodeAlreadyExists,
		fmt.Errorf("%s already exists: %s=%s", resource, field, value),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "RESOURCE_ALREADY_EXISTS",
		Domain: domain,
		Metadata: map[string]string{
			"resource": resource,
			"field":    field,
			"value":    value,
		},
	})
	err.AddDetail(detail)

	return err
}

// NewInternalError creates an INTERNAL error with ErrorInfo details.
//
// Usage in handlers:
//
//	result, err := h.service.CreateNote(ctx, params)
//	if err != nil {
//	    return nil, errors.NewInternalError(errors.MindDomain, "failed to create note", err)
//	}
func NewInternalError(domain, operation string, cause error) error {
	err := connect.NewError(
		connect.CodeInternal,
		fmt.Errorf("%s: %w", operation, cause),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "INTERNAL_ERROR",
		Domain: domain,
		Metadata: map[string]string{
			"operation": operation,
			"cause":     cause.Error(),
		},
	})
	err.AddDetail(detail)

	return err
}

// NewPermissionDeniedError creates a PERMISSION_DENIED error with ErrorInfo details.
//
// Usage in handlers:
//
//	if err == ErrCollectionIsSystem {
//	    return nil, errors.NewPermissionDeniedError(errors.MindDomain, "cannot delete system collection")
//	}
func NewPermissionDeniedError(domain, reason string) error {
	err := connect.NewError(
		connect.CodePermissionDenied,
		fmt.Errorf("permission denied: %s", reason),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "PERMISSION_DENIED",
		Domain: domain,
		Metadata: map[string]string{
			"reason": reason,
		},
	})
	err.AddDetail(detail)

	return err
}

// NewInvalidArgumentError creates an INVALID_ARGUMENT error with FieldViolation details.
//
// Usage in handlers:
//
//	if req.Msg.Title == "" {
//	    return nil, errors.NewInvalidArgumentError("title", "title cannot be empty")
//	}
func NewInvalidArgumentError(field, description string) error {
	err := connect.NewError(
		connect.CodeInvalidArgument,
		fmt.Errorf("invalid argument: %s - %s", field, description),
	)

	// Add BadRequest with FieldViolations per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       field,
				Description: description,
			},
		},
	})
	err.AddDetail(detail)

	return err
}

// NewFailedPreconditionError creates a FAILED_PRECONDITION error with ErrorInfo details.
// This is used for optimistic locking failures (ETag mismatches) and other precondition failures.
//
// Usage in handlers:
//
//	metadata := map[string]string{
//	    "provided_etag": ifMatchHeader,
//	    "current_etag":  currentETag,
//	}
//	return nil, errors.NewFailedPreconditionError(errors.MindDomain, "ETAG_MISMATCH", metadata)
func NewFailedPreconditionError(domain, reason string, metadata map[string]string) error {
	err := connect.NewError(
		connect.CodeFailedPrecondition,
		fmt.Errorf("precondition failed: %s", reason),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	})
	err.AddDetail(detail)

	return err
}
