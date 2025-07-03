package config

var (
	RouteHome                    = "/"
	RoutePrompts                 = "/prompts"
	RoutePromptsEdit             = "/prompts/edit/:id"
	RoutePromptsDelete           = "/prompts/delete"
	RouteProviders               = "/providers"
	RouteProvidersEdit           = "/providers/edit/:id"
	RouteProvidersDelete         = "/providers/delete"
	RouteLLMServices             = "/llm-services"
	RouteLLMServicesEdit         = "/llm-services/edit/:id"
	RouteLLMServicesDelete       = "/llm-services/delete"
	RouteLLMServicesModels       = "/llm-services/models"
	RouteLLMServiceConfigs       = "/llm-service-configs"
	RouteLLMServiceConfigsEdit   = "/llm-service-configs/edit/:id"
	RouteLLMServiceConfigsDelete = "/llm-service-configs/delete"
	RouteLLMServiceConfigsModels = "/llm-service-configs/models"
	RouteSettings                = "/settings"
	RouteConversations           = "/conversations"
	RouteConversationsNew        = "/conversations/new"
	RouteConversationsView       = "/conversations/:id"
	RouteConversationsCreate     = "/conversations/create"
	// Auth routes
	RouteAuthSignIn         = "/auth/signin"
	RouteAuthSignUp         = "/auth/signup"
	RouteAuthForgotPassword = "/auth/forgot-password"
	// Asset routes
	RouteAssets = "/assets"
)
