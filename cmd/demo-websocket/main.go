package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/internal/service"
	"go-demo/internal/ws"
	"go-demo/pkg/gox"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{} // use default options
var pongWait = 30 * time.Second     // 心跳超时
var pingPeriod = pongWait / 4       // 心跳间隔

func socketHandler(w http.ResponseWriter, r *http.Request) {
	client := &service.WsClient{Conn: nil, IsClosed: true}
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}
	client.Conn = conn
	client.IsClosed = false
	// Close
	defer service.Ws.Close(client)

	// 鉴权, jwt redis白名单
	clientId := r.URL.Query().Get("client_id") // url_base64(userId:jwtSignature)
	clientIdDecoded, err := base64.RawURLEncoding.DecodeString(clientId)
	if err != nil {
		service.Ws.Send(client, "ClientError", map[string]any{
			"code":    "UserUnauthorized",
			"message": "您未登录或登录已过期, 请重新登录",
		})
		return
	}
	userJwt := strings.Split(string(clientIdDecoded), ":")
	if len(userJwt) != 2 {
		service.Ws.Send(client, "ClientError", map[string]any{
			"code":    "UserUnauthorized",
			"message": "您未登录或登录已过期, 请重新登录",
		})
		return
	}
	key := fmt.Sprintf(consts.JwtLogin, consts.UserJwt, userJwt[0], userJwt[1])
	if n, err := di.JwtRedis().Exists(context.Background(), key).Result(); err != nil {
		service.Ws.Send(client, "ClientError", map[string]any{
			"code":    "InternalError",
			"message": "服务异常, 请稍后重试",
		})
		return
	} else if 0 == n {
		service.Ws.Send(client, "ClientError", map[string]any{
			"code":    "UserUnauthorized",
			"message": "您未登录或登录已过期, 请重新登录",
		})
		return
	}

	// 心跳
	if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		zap.L().Error(err.Error())
	}
	gox.Go(func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			if client.IsClosed {
				return
			}
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				zap.L().Error(err.Error())
			}
		}
	})
	client.Conn.SetPongHandler(func(appData string) error {
		if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			zap.L().Error(err.Error())
		}
		return nil
	})

	// 接收消息, 格式 {type: "", data: {}}
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.L().Error(err.Error())
			}
			break
		}
		msg := service.WsMsg{}
		if err := json.Unmarshal(message, &msg); err != nil {
			service.Ws.Send(client, "ClientError", map[string]any{
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
			service.Ws.Send(client, "ClientError", map[string]any{
				"code":    "TypeError",
				"message": "未知消息类型",
			})
		}
	}
}

func main() {
	http.HandleFunc("/websocket", socketHandler)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
