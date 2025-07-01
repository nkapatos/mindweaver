package router

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
	"github.com/nkapatos/mindweaver/internal/router/middleware"
	"github.com/nkapatos/mindweaver/internal/router/routes"
)

type Router struct {
	echo *echo.Echo
}

func New() *Router {
	e := echo.New()

	// Configure logger to only show errors and important info
	loggerConfig := echoMiddleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency}\n",
		Skipper: func(c echo.Context) bool {
			// Skip logging successful requests (status < 400)
			return c.Response().Status < 400
		},
	}

	// Global middleware
	e.Use(echoMiddleware.LoggerWithConfig(loggerConfig))
	e.Use(middleware.HTMXMiddleware())
	e.Use(echoMiddleware.Recover())

	return &Router{echo: e}
}

func (r *Router) SetupRoutes(
	actorHandler *api.ActorHandler,
	promptHandler *api.PromptHandler,
	llmHandler *api.LLMHandler,
	homeHandler *web.HomeHandler,
	promptsHandler *web.PromptsHandler,
	providersHandler *web.ProvidersHandler,
	llmServicesHandler *web.LLMServicesHandler,
	llmServiceConfigsHandler *web.LLMServiceConfigsHandler,
	settingsHandler *web.SettingsHandler,
	conversationHandler *web.ConversationHandler,
	notFoundHandler *web.NotFoundHandler,
) {
	routes.SetupWebRoutes(r.echo, homeHandler, promptsHandler, providersHandler, llmServicesHandler, llmServiceConfigsHandler, settingsHandler, conversationHandler)
	if llmHandler != nil {
		routes.SetupAPIRoutes(r.echo, actorHandler, promptHandler, llmHandler)
	} else {
		routes.SetupAPIRoutes(r.echo, actorHandler, promptHandler, nil)
	}
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
