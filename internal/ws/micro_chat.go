package ws

import (
	"go-demo/internal/service"

	"github.com/spf13/cast"
)

type microChat struct{}

var MicroChat microChat

func (microChat) SendMessage(client *service.WsClient, data map[string]any) {
	content := cast.ToString(data["content"])
	if "" == content {
		return
	}
	content = "yes, " + content

	_ = service.Ws.Send(client, "MicroChat:SendMessage", map[string]any{
		"content": content,
	})
}
