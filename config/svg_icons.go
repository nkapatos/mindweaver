package config

import "fmt"

type SvgIconEnum int

const (
	IconHome SvgIconEnum = iota
	IconPrompts
	IconLLMServices
	IconConfigurations
	IconProviders
	IconSettings
	IconConversations
	IconMenu
	IconAttachment
	IconSend
	IconDownArrow
	IconLeftArrow
	IconMsgBubble
	IconMoreOptsVertical
)

var SvgIcons = map[SvgIconEnum]string{
	IconHome:             "house",
	IconPrompts:          "terminal",
	IconLLMServices:      "zap",
	IconConfigurations:   "settings-2",
	IconProviders:        "server",
	IconSettings:         "settings",
	IconConversations:    "message-square-more",
	IconMenu:             "menu",
	IconAttachment:       "paperclip",
	IconSend:             "send",
	IconDownArrow:        "chevron-down",
	IconLeftArrow:        "chevron-left",
	IconMsgBubble:        "message-circle",
	IconMoreOptsVertical: "ellipsis-vertical",
}

func GetSvgIconPath(icon SvgIconEnum) string {
	return fmt.Sprintf("static/icons/%s.svg", SvgIcons[icon])
}
