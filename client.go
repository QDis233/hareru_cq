package hareru_cq

import (
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

// Client cqhttp客户端
type Client struct {
	WsUrl             string
	AccessToken       string
	EnableAccessToken bool

	ActConn     *websocket.Conn
	EventConn   *websocket.Conn
	initialized bool
}

// CqRequest cqhttp请求
type CqRequest struct {
	Action string `json:"action"`
	Params any    `json:"params"`
	Echo   string `json:"echo"`
}

// CqResponse cqhttp响应
type CqResponse struct {
	Status  string `json:"status"`
	RetCode int    `json:"retcode"`
	Msg     string `json:"msg"`
	Wording string `json:"wording"`
	Echo    string `json:"echo"`

	Json gjson.Result
}

// Event 事件
type Event struct {
	Time    int64  `json:"time"`
	Type    string `json:"post_type"`
	SubType string `json:"sub_type"`

	Json gjson.Result
}

// Get Json取 Event 数据
func (event *Event) Get(path string) gjson.Result {
	return event.Json.Get(path)
}

// NewClient 创建 Client
func NewClient(wsUrl string, accessToken string) *Client {
	var enableToken bool
	if accessToken != "" {
		enableToken = true
	}

	return &Client{
		WsUrl:             wsUrl,
		AccessToken:       accessToken,
		EnableAccessToken: enableToken,
	}
}

// connect 连接 websocket API
func (c *Client) connect() error {
	actConn, _, err := websocket.DefaultDialer.Dial(c.WsUrl+"/api", nil)
	if err != nil {
		log.Fatal("WebSocket连接失败:", err)
		return err
	}

	eventConn, _, err := websocket.DefaultDialer.Dial(c.WsUrl+"/event", nil)
	if err != nil {
		log.Fatal("WebSocket连接失败:", err)
		return err
	}

	c.ActConn = actConn
	c.EventConn = eventConn

	log.Infoln("cqHttp 连接成功")

	return nil
}

// Close 关闭连接
func (c *Client) Close() {
	_ = c.ActConn.Close()
	_ = c.EventConn.Close()
	c.initialized = false
}

// Init 初始化
func (c *Client) Init() error {
	if c.initialized {
		log.Infoln("Client 已初始化")
		return &AlreadyInitializedErr{
			Message: "Client 已初始化",
		}
	}

	err := c.connect()
	if err != nil {
		log.Fatal("Client 初始化失败:", err)
		return err
	}

	c.initialized = true
	return nil
}

// IsInitialized 初始化状态
func (c *Client) IsInitialized() bool {
	return c.initialized
}
