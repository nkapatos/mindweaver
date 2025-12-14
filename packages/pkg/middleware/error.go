package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/pkg/types"
)

// ErrorHandlerMiddleware standardizes error responses for the API.
// This middleware converts echo.HTTPError to the standardized types.Response format with AIP-193 compliance.
func ErrorHandlerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			code := http.StatusInternalServerError
			msg := http.StatusText(code)
			var errorResp *types.ErrorResponse

			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code

				// Check if Message is already a types.ErrorResponse (structured error)
				if errResp, ok := he.Message.(*types.ErrorResponse); ok {
					errorResp = errResp
				} else if errResp, ok := he.Message.(types.ErrorResponse); ok {
					errorResp = &errResp
				} else if m, ok := he.Message.(string); ok {
					// Simple string message
					msg = m
					errorResp = &types.ErrorResponse{
						Code:    code,
						Message: msg,
						Status:  httpStatusToAIPStatus(code),
					}
				} else {
					// Fallback
					msg = http.StatusText(code)
					errorResp = &types.ErrorResponse{
						Code:    code,
						Message: msg,
						Status:  httpStatusToAIPStatus(code),
					}
				}
			} else {
				// Not an echo.HTTPError, generic error
				errorResp = &types.ErrorResponse{
					Code:    code,
					Message: err.Error(),
					Status:  httpStatusToAIPStatus(code),
				}
			}

			resp := types.Response[any]{
				Error: errorResp,
			}
			return c.JSON(code, resp)
		}
		return nil
	}
}

// httpStatusToAIPStatus maps HTTP status codes to AIP-193 status strings.
// Reference: https://google.aip.dev/193 and https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func httpStatusToAIPStatus(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "INVALID_ARGUMENT"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "ALREADY_EXISTS"
	case http.StatusForbidden:
		return "PERMISSION_DENIED"
	case http.StatusUnauthorized:
		return "UNAUTHENTICATED"
	case http.StatusServiceUnavailable:
		return "UNAVAILABLE"
	case http.StatusInternalServerError:
		return "INTERNAL"
	case http.StatusPreconditionFailed: // 412
		return "FAILED_PRECONDITION"
	case http.StatusRequestEntityTooLarge: // 413
		return "OUT_OF_RANGE"
	case http.StatusTooManyRequests: // 429
		return "RESOURCE_EXHAUSTED"
	case http.StatusNotImplemented: // 501
		return "UNIMPLEMENTED"
	case 410: // Gone
		return "DATA_LOSS"
	case http.StatusRequestTimeout: // 408
		return "ABORTED"
	default:
		return "UNKNOWN"
	}
}
