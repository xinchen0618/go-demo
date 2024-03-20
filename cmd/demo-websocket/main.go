// WebSocket 入口
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-demo/config/di"
	"go-demo/internal/consts"
	"go-demo/internal/service"
	"go-demo/internal/types"
	"go-demo/internal/ws"
	"go-demo/pkg/gox"

	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

var (
	upgrader   = websocket.Upgrader{} // use default options
	pongWait   = 30 * time.Second     // 心跳超时
	pingPeriod = pongWait / 4         // 心跳间隔
)

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// 将 ws 连接信息和 user_id 记录到 WSClient 对象
	client := &types.WSClient{Conn: nil, IsClosed: true}

	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		di.Logger().Error(err.Error())
		return
	}
	client.Conn = conn
	client.IsClosed = false
	// Close
	defer service.WS.Close(client)

	// 鉴权, jwt redis 白名单
	clientID := r.URL.Query().Get("client_id") // url_base64(userID:md5(jwtToken))
	clientIDDecoded, err := base64.RawURLEncoding.DecodeString(clientID)
	if err != nil {
		_ = service.WS.Send(client, "ClientError", map[string]any{
			"code":    "UserUnauthorized",
			"message": "您未登录或登录已过期, 请重新登录",
		})
		return
	}
	userJWT := strings.Split(string(clientIDDecoded), ":")
	if len(userJWT) != 2 {
		_ = service.WS.Send(client, "ClientError", map[string]any{
			"code":    "UserUnauthorized",
			"message": "您未登录或登录已过期, 请重新登录",
		})
		return
	}
	key := fmt.Sprintf(consts.JWTLogin, consts.UserJWT, userJWT[0], userJWT[1])
	if n, err := di.JWTRedis().Exists(context.Background(), key).Result(); err != nil {
		_ = service.WS.Send(client, "ClientError", map[string]any{
			"code":    "InternalError",
			"message": "服务异常, 请稍后重试",
		})
		return
	} else if n == 0 {
		_ = service.WS.Send(client, "ClientError", map[string]any{
			"code":    "UserUnauthorized",
			"message": "您未登录或登录已过期, 请重新登录",
		})
		return
	}
	client.UserID = cast.ToInt64(userJWT[0])

	// 心跳
	if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		di.Logger().Error(err.Error())
	}
	gox.SafeGo(func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			if client.IsClosed {
				return
			}
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				di.Logger().Error(err.Error())
			}
		}
	})
	client.Conn.SetPongHandler(func(appData string) error {
		if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			di.Logger().Error(err.Error())
		}
		return nil
	})

	/************************* 服务端主动向客户端推送消息 **************************/
	// 这里通过 redis 订阅来实现, 服务端监听名为 wsMessageChannel 的 redis 频道
	// 向频道发送消息的格式为 json 字符串 `{"user_id": int, "type": string, data: {}}`
	// user_id 为 0 表示向所有用户推送消息, 否则为向指定用户推送消息
	pubsub := di.StorageRedis().Subscribe(context.Background(), "wsMessageChannel") // 订阅一个或多个频道
	// 检查订阅是否成功
	if _, err := pubsub.Receive(context.Background()); err != nil {
		di.Logger().Error(err.Error())
		_ = service.WS.Send(client, "InternalError", map[string]any{ // 订阅失败
			"code":    "InternalError",
			"message": "服务异常, 请稍后重试",
		})
		return
	}
	// 创建一个通道来接收订阅的消息
	msgCh := pubsub.Channel()
	// 启动一个 goroutine 来处理订阅的消息
	gox.SafeGo(func() {
		for msg := range msgCh {
			// 读取订阅消息并格式化
			submsg := types.SubMsg{}
			if err := json.Unmarshal([]byte(msg.Payload), &submsg); err != nil {
				di.Logger().Error(err.Error())
				continue
			}
			if submsg.UserID != 0 && submsg.UserID != client.UserID { // 并非当前客户端的消息
				continue
			}

			// 业务路由
			switch submsg.Type {
			case "MicroChat:SendMessage": // DEMO
				ws.MicroChat.SendMessage(client, submsg.Data)
			default: // 未知路由
				di.Logger().Error(fmt.Sprintf("ws 错误订阅消息: %s", msg.Payload))
			}
		}
	})

	/************************* 服务端接收客户端发来的消息 **************************/
	// 消息格式为 json 字符串 `{type: "", data: {}}`
	for {
		// 读取客户端发送的消息并格式化
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				di.Logger().Error(err.Error())
			}
			break
		}
		msg := types.WSMsg{}
		if err := json.Unmarshal(message, &msg); err != nil {
			_ = service.WS.Send(client, "ClientError", map[string]any{
				"code":    "MessageError",
				"message": "消息格式不正确",
			})
			continue
		}

		// 业务路由
		switch msg.Type {
		case "MicroChat:SendMessage": // DEMO
			ws.MicroChat.SendMessage(client, msg.Data)
		default: // 未知路由
			_ = service.WS.Send(client, "ClientError", map[string]any{
				"code":    "TypeError",
				"message": "未知消息类型",
			})
		}
	}
}

func main() {
	http.HandleFunc("/websocket", socketHandler)
	if err := http.ListenAndServe(":9090", nil); err != nil {
		di.Logger().Error(err.Error())
		return
	}
}
