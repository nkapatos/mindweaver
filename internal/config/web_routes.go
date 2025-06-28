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
		Href: "/",
		Icon: "house",
		Text: "Home",
	},
	{
		Href: "/prompts",
		Icon: "terminal",
		Text: "Prompts",
	},
	{
		Href: "/llm-services",
		Icon: "zap",
		Text: "LLM Services",
	},
	{
		Href: "/providers",
		Icon: "server",
		Text: "Providers",
	},
	{
		Href: "/settings",
		Icon: "settings",
		Text: "Settings",
	},
	{
		Href: "/chats",
		Icon: "message-square-more",
		Text: "Chats",
	},
}
