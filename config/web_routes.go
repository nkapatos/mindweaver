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
		Icon: GetSvgIconPath(IconHome),
		Text: "Home",
	},
	{
		Href: RoutePrompts,
		Icon: GetSvgIconPath(IconPrompts),
		Text: "Prompts",
	},
	{
		Href: RouteLLMServices,
		Icon: GetSvgIconPath(IconLLMServices),
		Text: "LLM Services",
	},
	{
		Href: RouteLLMServiceConfigs,
		Icon: GetSvgIconPath(IconConfigurations),
		Text: "Configurations",
	},
	{
		Href: RouteProviders,
		Icon: GetSvgIconPath(IconProviders),
		Text: "Providers",
	},
	{
		Href: RouteSettings,
		Icon: GetSvgIconPath(IconSettings),
		Text: "Settings",
	},
	{
		Href: RouteConversations,
		Icon: GetSvgIconPath(IconConversations),
		Text: "Conversations",
	},
}
