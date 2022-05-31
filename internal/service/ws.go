package service

import (
	"encoding/json"
	"errors"

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

var clientHub = map[*WsClient]struct{}{}

// SetClientHub
//  @receiver ws
//  @param client *WsClient
func (ws) SetClientHub(client *WsClient) {
	clientHub[client] = struct{}{}
}

// DeleteClientHub
//  @receiver ws
//  @param client *WsClient
func (ws) DeleteClientHub(client *WsClient) {
	delete(clientHub, client)
}

// Send 发送消息
//  @receiver ws
//  @param client *WsClient
//  @param msgType string
//  @param msgData map[string]any
//  @return error
func (ws) Send(client *WsClient, msgType string, msgData map[string]any) error {
	if client.IsClosed {
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
		zap.L().Error(err.Error())
		return err
	}
	if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		zap.L().Error(err.Error())
		return err
	}
	return nil
}

// Broadcast 广播消息
//  @receiver ws
//  @param msgType string
//  @param msgData map[string]any
func (ws) Broadcast(msgType string, msgData map[string]any) {
	for client := range clientHub {
		if err := Ws.Send(client, msgType, msgData); err != nil {
			zap.L().Error(err.Error())
		}
	}
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
	Ws.DeleteClientHub(client)
}
