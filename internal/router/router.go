package router

import (
	"crypto/subtle"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
	"github.com/nkapatos/mindweaver/internal/router/middleware"
	"github.com/nkapatos/mindweaver/internal/router/routes"
	"github.com/nkapatos/mindweaver/internal/services"
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
	e.Use(middleware.RouterPathMiddleware())
	e.Use(echoMiddleware.Recover())

	return &Router{echo: e}
}

func (r *Router) SetupRoutes(
	authService *services.AuthService,
	authHandler *web.AuthHandler,
	actorHandler *api.ActorHandler,
	promptHandler *api.PromptHandler,
	llmHandler *api.LLMHandler,
	conversationHandler *api.ConversationHandler,
	providerHandler *api.ProvidersHandler,
	llmServiceHandler *api.LLMServicesHandler,
	llmServiceConfigHandler *api.LLMServiceConfigsHandler,
	modelsHandler *api.ModelsHandler,
	homeHandler *web.HomeHandler,
	promptsHandler *web.PromptsHandler,
	providersHandler *web.ProvidersHandler,
	llmServicesHandler *web.LLMServicesHandler,
	llmServiceConfigsHandler *web.LLMServiceConfigsHandler,
	settingsHandler *web.SettingsHandler,
	webConversationHandler *web.ConversationHandler,
	notFoundHandler *web.NotFoundHandler,
	setupHandler *web.SetupHandler,
) {
	e := r.echo

	// Add session middleware using Echo's session middleware with gorilla/sessions
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("mindweaver-secret-key-change-in-production"))))

	// TODO: Replace basic auth with proper session-based authentication
	// For now, we'll use basic auth as a placeholder
	basicAuthMiddleware := echoMiddleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// TODO: Use authService.AuthenticateActor() for proper authentication
		// For now, use hardcoded credentials for testing
		if subtle.ConstantTimeCompare([]byte(username), []byte("testuser")) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte("testpass123")) == 1 {
			return true, nil
		}
		return false, nil
	})

	routes.SetupWebRoutes(e, authHandler, basicAuthMiddleware, homeHandler, promptsHandler, providersHandler, llmServicesHandler, llmServiceConfigsHandler, settingsHandler, webConversationHandler, setupHandler)
	if llmHandler != nil {
		routes.SetupAPIRoutes(e, actorHandler, promptHandler, llmHandler, conversationHandler, providerHandler, llmServiceHandler, llmServiceConfigHandler, modelsHandler)
	} else {
		routes.SetupAPIRoutes(e, actorHandler, promptHandler, nil, conversationHandler, providerHandler, llmServiceHandler, llmServiceConfigHandler, modelsHandler)
	}
	routes.SetupStaticRoutes(e)
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
