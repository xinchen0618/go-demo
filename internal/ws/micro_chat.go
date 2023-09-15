// Package ws websocket 服务
package ws

import (
	"go-demo/internal/service"
	"go-demo/internal/types"

	"github.com/spf13/cast"
)

type microChat struct{}

var MicroChat microChat

func (microChat) SendMessage(client *types.WSClient, data map[string]any) {
	content := cast.ToString(data["content"])
	if content == "" {
		return
	}
	content = "yes, " + content

	_ = service.WS.Send(client, "MicroChat:SendMessage", map[string]any{
		"content": content,
	})
}
