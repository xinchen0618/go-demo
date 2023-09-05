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

// WSClient 客户端信息
//
//	这个对象用于存放客户端的 ws 连接信息和用户信息
type WSClient struct {
	UserID   int64
	Conn     *websocket.Conn
	IsClosed bool
}

// WSMsg 客户端与服务端通信的消息格式
type WSMsg struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// SubMsg 服务端订阅 redis 频道的消息格式
type SubMsg struct {
	UserID int64          `json:"user_id"`
	Type   string         `json:"type"`
	Data   map[string]any `json:"data"`
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

// Close 关闭 client
func (ws) Close(client *WSClient) {
	if client.IsClosed {
		return
	}
	if err := client.Conn.Close(); err != nil {
		di.Logger().Error(err.Error())
	}
	client.IsClosed = true
}
