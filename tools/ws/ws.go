package ws

import (
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Upgrade 升级http协议为websocket协议
func Upgrade(c *gin.Context) (conn *websocket.Conn, err error) {
	wsUpgrade := websocket.Upgrader{
		ReadBufferSize:  10240, // 读缓冲区大小
		WriteBufferSize: 10240, // 写缓冲区大小
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return wsUpgrade.Upgrade(c.Writer, c.Request, nil)
}

// New 创建websocket连接
func New(url string) (conn *websocket.Conn, err error) {
	err = retry.Do(func() error {
		conn, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			zap.L().Info("websocket 连接失败, 重试连接", zap.Error(err))
			return err
		}
		return nil
	},
		retry.Attempts(3),
		retry.Delay(1*time.Second),
	)
	if err != nil {
		zap.L().Error("websocket 连接失败", zap.Error(err))
		return
	}
	zap.L().Info("websocket 连接成功")
	return
}

// Close 关闭服务端连接
func Close(conns ...*websocket.Conn) {
	for _, conn := range conns {
		err := conn.Close()
		if err != nil {
			zap.L().Error("websocket 连接关闭失败", zap.Error(err))
		}
		zap.L().Info("websocket 连接关闭成功")
	}
}

// Write 发送文本消息
func Write(conn *websocket.Conn, msg ...[]byte) {
	for _, v := range msg {
		err := conn.WriteMessage(websocket.TextMessage, v)
		if err != nil {
			zap.L().Error("websocket 发送文本消息失败", zap.Error(err))
			Close(conn)
		}
	}
}

// WriteOnce 发送文本消息 (只发送一次)
func WriteOnce(conn *websocket.Conn, msg []byte) {
	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		zap.L().Error("websocket 发送文本消息失败", zap.Error(err))
	}
	Close(conn)
}
