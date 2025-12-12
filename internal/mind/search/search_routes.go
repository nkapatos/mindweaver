package search

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers the search routes with the Echo router.
// Routes:
//
//	GET /api/mind/search - Full-text search on notes
func RegisterRoutes(g *echo.Group, handler *SearchHandler) {
	g.GET("", handler.SearchNotesHandler)
}
