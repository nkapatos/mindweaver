// V3 Error Helpers - Connect-RPC errors with proper details
package collections

import (
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

const (
	// Domain for ErrorInfo per Google standards
	errorDomain = "mind.mindweaver.com"
)

// newAlreadyExistsError creates an ALREADY_EXISTS error with ErrorInfo details
func newAlreadyExistsError(resource, field, value string) error {
	err := connect.NewError(
		connect.CodeAlreadyExists,
		fmt.Errorf("%s already exists: %s=%s", resource, field, value),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "RESOURCE_ALREADY_EXISTS",
		Domain: errorDomain,
		Metadata: map[string]string{
			"resource": resource,
			"field":    field,
			"value":    value,
		},
	})
	err.AddDetail(detail)

	return err
}

// newNotFoundError creates a NOT_FOUND error with ErrorInfo details
func newNotFoundError(resource, resourceID string) error {
	err := connect.NewError(
		connect.CodeNotFound,
		fmt.Errorf("%s not found: %s", resource, resourceID),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "RESOURCE_NOT_FOUND",
		Domain: errorDomain,
		Metadata: map[string]string{
			"resource": resource,
			"id":       resourceID,
		},
	})
	err.AddDetail(detail)

	return err
}

// newPermissionDeniedError creates a PERMISSION_DENIED error with ErrorInfo details
func newPermissionDeniedError(reason string) error {
	err := connect.NewError(
		connect.CodePermissionDenied,
		fmt.Errorf("permission denied: %s", reason),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "PERMISSION_DENIED",
		Domain: errorDomain,
		Metadata: map[string]string{
			"reason": reason,
		},
	})
	err.AddDetail(detail)

	return err
}

// newInternalError creates an INTERNAL error with ErrorInfo details
func newInternalError(operation string, cause error) error {
	err := connect.NewError(
		connect.CodeInternal,
		fmt.Errorf("%s: %w", operation, cause),
	)

	// Add ErrorInfo details per AIP-193
	detail, _ := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "INTERNAL_ERROR",
		Domain: errorDomain,
		Metadata: map[string]string{
			"operation": operation,
			"cause":     cause.Error(),
		},
	})
	err.AddDetail(detail)

	return err
}

// newInvalidArgumentError creates an INVALID_ARGUMENT error with FieldViolation details
func newInvalidArgumentError(field, description string) error {
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
