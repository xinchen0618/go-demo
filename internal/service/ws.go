// Package service 内部应用业务原子级服务
//
//	需要公共使用的业务逻辑在这里实现.
package service

import (
	"errors"
	"fmt"

	"go-demo/config/di"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type WSClient struct {
	Conn     *websocket.Conn
	IsClosed bool
}

// WSMsg 客户端与服务器通信的消息格式
type WSMsg struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type ws struct{}

var WS ws

// Send 发送消息
func (ws) Send(client *WSClient, msgType string, msgData map[string]any) error {
	if client.IsClosed {
		di.Logger().Error(fmt.Sprintf("%p client is closed", client))
		return errors.New("client is closed")
	}

	if msgData == nil {
		msgData = map[string]any{}
	}
	message, err := json.Marshal(WSMsg{
		Type: msgType,
		Data: msgData,
	})
	if err != nil {
		di.Logger().Error(err.Error())
		return err
	}
	if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		di.Logger().Error(err.Error())
		return err
	}

	return nil
}

// Close 关闭client
func (ws) Close(client *WSClient) {
	if client.IsClosed {
		return
	}
	if err := client.Conn.Close(); err != nil {
		di.Logger().Error(err.Error())
	}
	client.IsClosed = true
}
