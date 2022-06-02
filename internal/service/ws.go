package service

import (
	"encoding/json"
	"fmt"

	"go-demo/config/di"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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
//  方法为异步处理
//  @receiver ws
//  @param client *WsClient
//  @param msgType string
//  @param msgData map[string]any
//  @return error
func (ws) Send(client *WsClient, msgType string, msgData map[string]any) {
	if client.IsClosed {
		zap.L().Error(fmt.Sprintf("%p client is closed", client))
		return
	}

	di.WorkerPool().Submit(func() {
		if nil == msgData {
			msgData = map[string]any{}
		}
		message, err := json.Marshal(WsMsg{
			Type: msgType,
			Data: msgData,
		})
		if err != nil {
			zap.L().Error(err.Error())
			return
		}
		if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			zap.L().Error(err.Error())
		}
	})
}

// Close 关闭client
//  @receiver ws
//  @param client *WsClient
func (ws) Close(client *WsClient) {
	if client.IsClosed {
		return
	}
	if err := client.Conn.Close(); err != nil {
		zap.L().Error(err.Error())
	}
	client.IsClosed = true
}
