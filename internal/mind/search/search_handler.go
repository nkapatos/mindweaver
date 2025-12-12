// TODO: Future search endpoint enhancements:
// - Support filter params: ?meta=key:value, ?tags=tag1,tag2, ?collection_id=123
// - Support sort params: ?sort=relevance|created|title
// - Support faceted search results (group by collection, type, tags)
// ============================================================================

package search

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// SearchHandler provides HTTP handlers for full-text search.
type SearchHandler struct {
	searchSvc *SearchService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(service *SearchService) *SearchHandler {
	return &SearchHandler{
		searchSvc: service,
	}
}

// SearchNotesHandler handles GET /api/mind/search
// Query params: q (required), limit, offset, include_body, min_score
func (h *SearchHandler) SearchNotesHandler(c echo.Context) error {
	// Parse query parameters
	var req SearchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query parameters")
	}

	// Validate required query
	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}

	// Convert to service query
	serviceQuery := ToServiceQuery(req)

	// Execute search
	resp, err := h.searchSvc.Search(c.Request().Context(), serviceQuery)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "search failed").SetInternal(err)
	}

	// Convert to API response
	apiResp := ToAPIResponse(resp, int(serviceQuery.Limit), int(serviceQuery.Offset))

	return c.JSON(http.StatusOK, apiResp)
}
