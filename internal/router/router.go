package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
	"github.com/nkapatos/mindweaver/internal/router/routes"
)

type Router struct {
	echo *echo.Echo
}

func New() *Router {
	e := echo.New()

	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return &Router{echo: e}
}

func (r *Router) SetupRoutes(
	actorHandler *api.ActorHandler,
	promptHandler *api.PromptHandler,
	llmHandler *api.LLMHandler,
	homeHandler *web.HomeHandler,
	promptsHandler *web.PromptsHandler,
	providersHandler *web.ProvidersHandler,
	settingsHandler *web.SettingsHandler,
	conversationHandler *web.ConversationHandler,
	notFoundHandler *web.NotFoundHandler,
) {
	routes.SetupWebRoutes(r.echo, homeHandler, promptsHandler, providersHandler, settingsHandler, conversationHandler)
	routes.SetupAPIRoutes(r.echo, actorHandler, promptHandler, llmHandler)
	routes.SetupStaticRoutes(r.echo)
	r.setupErrorHandling(notFoundHandler)
}

func (r *Router) setupErrorHandling(notFoundHandler *web.NotFoundHandler) {
	// 404 handler
	r.echo.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok && he.Code == 404 {
			notFoundHandler.NotFound(c)
			return
		}
		r.echo.DefaultHTTPErrorHandler(err, c)
	}
}

func (r *Router) Echo() *echo.Echo {
	return r.echo
}
