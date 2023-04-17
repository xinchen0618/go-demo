package service

import (
	"errors"
	"fmt"

	"go-demo/config/di"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
)

type WsClient struct {
	Conn     *websocket.Conn
	IsClosed bool
}

type WsMsg struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type ws struct{}

var Ws ws

// Send 发送消息
//
//	@receiver ws
//	@param client *WsClient
//	@param msgType string
//	@param msgData map[string]any
//	@return error
func (ws) Send(client *WsClient, msgType string, msgData map[string]any) error {
	if client.IsClosed {
		di.Logger().Error(fmt.Sprintf("%p client is closed", client))
		return errors.New("client is closed")
	}

	if nil == msgData {
		msgData = map[string]any{}
	}
	message, err := json.Marshal(WsMsg{
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
//
//	@receiver ws
//	@param client *WsClient
func (ws) Close(client *WsClient) {
	if client.IsClosed {
		return
	}
	if err := client.Conn.Close(); err != nil {
		di.Logger().Error(err.Error())
	}
	client.IsClosed = true
}
