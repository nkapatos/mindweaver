package config

// RouteInfo contains complete information for a web route
type RouteInfo struct {
	Href string
	Icon string
	Text string
}

// WebRoutes holds all web route information as a slice for iteration
var WebRoutes = []RouteInfo{
	{
		Href: RouteHome,
		Icon: "house",
		Text: "Home",
	},
	{
		Href: RoutePrompts,
		Icon: "terminal",
		Text: "Prompts",
	},
	{
		Href: RouteLLMServices,
		Icon: "zap",
		Text: "LLM Services",
	},
	{
		Href: RouteLLMServiceConfigs,
		Icon: "settings-2",
		Text: "Configurations",
	},
	{
		Href: RouteProviders,
		Icon: "server",
		Text: "Providers",
	},
	{
		Href: RouteSettings,
		Icon: "settings",
		Text: "Settings",
	},
	{
		Href: RouteConversations,
		Icon: "message-square-more",
		Text: "Conversations",
	},
}
