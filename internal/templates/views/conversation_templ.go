// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.898
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"encoding/json"

	"github.com/nkapatos/mindweaver/config"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/nkapatos/mindweaver/internal/templates/components"
	"github.com/nkapatos/mindweaver/internal/templates/elements"
	"github.com/nkapatos/mindweaver/internal/templates/layouts"
)

// ConversationMetadata represents the metadata structure for conversations
type ConversationMetadata struct {
	DefaultProviderID   int64  `json:"default_provider_id"`
	DefaultProviderName string `json:"default_provider_name"`
	ConversationType    string `json:"conversation_type,omitempty"`
}

// getDefaultProviderFromMetadata extracts provider info from conversation metadata
func getDefaultProviderFromMetadata(conversation *store.Conversation) *store.Provider {
	if conversation == nil || !conversation.Metadata.Valid {
		return nil
	}

	var metadata ConversationMetadata
	if err := json.Unmarshal([]byte(conversation.Metadata.String), &metadata); err != nil {
		return nil
	}

	// Create a provider struct from metadata
	return &store.Provider{
		ID:   metadata.DefaultProviderID,
		Name: metadata.DefaultProviderName,
	}
}

func Conversation(providerDropdownData components.ProviderDropdownData, messageInputData components.MessageInputData) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div class=\"flex flex-col bg-base-200\"><!-- Smart Chat Header - Fixed height --><div class=\"bg-base-100 border-b border-base-300 px-6 py-4 flex-shrink-0\"><div class=\"flex items-center justify-between\"><div class=\"flex items-center gap-4 flex-1\"><!-- Provider Selection Dropdown -->")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = components.ProviderDropdown(providerDropdownData).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<!-- Conversation Title/Info --><div class=\"flex-1 min-w-0\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if providerDropdownData.SelectedProvider != nil {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "<h1 class=\"text-lg font-semibold text-base-content\">New Conversation</h1><p class=\"text-sm text-base-content/60\">Using ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var3 string
				templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(providerDropdownData.SelectedProvider.Name)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/views/conversation.templ`, Line: 51, Col: 98}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, " - Ready to chat</p>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "<h1 class=\"text-lg font-semibold text-base-content\">New Conversation</h1><p class=\"text-sm text-base-content/60\">Choose a provider to start chatting</p>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "</div></div><!-- Action Buttons --><div class=\"flex items-center gap-2\"><button class=\"btn btn-ghost btn-sm\" title=\"More options\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = elements.Icon(config.GetSvgIconPath(config.IconMoreOptsVertical)).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "</button></div></div></div><!-- Messages Area --><div class=\"flex-1 overflow-y-auto p-6\" id=\"messages-container\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if providerDropdownData.SelectedProvider != nil {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, "<!-- Ready to chat state --> <div class=\"text-center text-base-content/60 py-12\"><div class=\"max-w-md mx-auto\"><div class=\"w-16 h-16 bg-primary rounded-full flex items-center justify-center mx-auto mb-4\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = elements.Icon(config.GetSvgIconPath(config.IconMsgBubble)).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "</div><h3 class=\"text-lg font-semibold mb-2\">Ready to Chat!</h3><p class=\"text-sm mb-6\">You're using <strong>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var4 string
				templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(providerDropdownData.SelectedProvider.Name)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/views/conversation.templ`, Line: 76, Col: 96}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "</strong>. Type your message below to start the conversation.</p><div class=\"space-y-2 text-xs text-base-content/50\"><p>💡 You can switch providers anytime using the dropdown</p><p>💡 Use @ mentions to switch providers mid-conversation</p></div></div></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, "<!-- Welcome/Empty State --> <div class=\"text-center text-base-content/60 py-12\"><div class=\"max-w-md mx-auto\"><div class=\"w-16 h-16 bg-base-300 rounded-full flex items-center justify-center mx-auto mb-4\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = elements.Icon(config.GetSvgIconPath(config.IconMsgBubble)).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, "</div><h3 class=\"text-lg font-semibold mb-2\">Start a New Conversation</h3><p class=\"text-sm mb-6\">Select a provider from the dropdown above to begin chatting with your AI assistant.</p><div class=\"space-y-2 text-xs text-base-content/50\"><p>💡 You can switch providers mid-conversation using @ mentions</p><p>💡 Each provider has different capabilities and costs</p></div></div></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, "</div><!-- Input Area - Fixed height -->")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = components.MessageInput(messageInputData).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, "</div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = layouts.AppLayout("Conversation", "Chat with your AI assistant").Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func NewConversationPage(providers []store.Provider, activePath string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var5 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var5 == nil {
			templ_7745c5c3_Var5 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = Conversation(components.ProviderDropdownData{
			Providers:        providers,
			SelectedProvider: nil,
			DropdownID:       "new-conversation-provider-dropdown",
		}, components.MessageInputData{
			Placeholder: "Type your message here...",
			IsDisabled:  true,
		}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
