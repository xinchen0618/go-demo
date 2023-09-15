// Package types 业务相关结构体定义
package types

import "github.com/gorilla/websocket"

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
