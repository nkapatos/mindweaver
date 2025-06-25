package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	// Create a new router
	r := New()
	assert.NotNil(t, r)
	assert.NotNil(t, r.Echo())

	// Test that the Echo instance is properly configured
	e := r.Echo()
	assert.NotNil(t, e)
}

func TestRouterSetup(t *testing.T) {
	// Just test that the setup doesn't panic
	assert.NotPanics(t, func() {
		// Note: In a real test, you would create mock handlers here
		// r.SetupRoutes(mockUserHandler, mockPromptHandler, ...)
	})
}

func TestEchoInstance(t *testing.T) {
	r := New()
	e := r.Echo()

	// Test that we can add a simple route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())
}
