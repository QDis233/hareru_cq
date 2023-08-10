package hareru_cq

import (
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"log"

	"github.com/gorilla/websocket"
)

const (
	ReceiveMessageEvent = "message"         // EventTypes
	PrivateMessageEvent = "private_message" // MessageEvents
	GroupMessageEvent   = "group_message"

	FriendRecallEvent  = "friend_recall" // NoticeEvents
	GroupRecallEvent   = "group_recall"
	GroupIncreaseEvent = "group_increase"
	GroupDecreaseEvent = "group_decrease"
	GroupAdminEvent    = "group_admin"
	GroupUploadEvent   = "group_upload"
	GroupBanEvent      = "group_ban"
	FriendAddEvent     = "friend_add"
	GroupCardEvent     = "group_card"
	EssenceEvent       = "essence"

	FriendRequestEvent = "friend_request" //RequestEvents
	GroupRequestEvent  = "group_request"

	LifecycleEvent = "lifecycle" //MetaEvents
	HeartbeatEvent = "heartbeat"
)

type Bot struct {
	Client *Client

	Info        *BotInfo
	ResChan     map[string]chan *CqResponse
	initialized bool
}

// BotInfo bot信息
type BotInfo struct {
	UserId   int64  //QQ
	NickName string //昵称
}

func (bot *Bot) doAction(req *CqRequest) error {
	err := bot.Client.ActConn.WriteJSON(req)
	if err != nil {
		log.Fatal("发送请求失败:", err)
		return err
	}
	return nil
}

func (bot *Bot) getActionResult(echo string) *CqResponse {
	resChan := make(chan *CqResponse)

	bot.ResChan[echo] = resChan

	res := <-resChan

	delete(bot.ResChan, echo)
	bot.ResChan[echo] = nil
	return res
}

func (bot *Bot) Stop() {
	bot.Client.Close()
}

func (bot *Bot) Init() error {
	if !bot.Client.IsInitialized() {
		err := bot.Client.Init()
		if err != nil {
			return err
		}
	}

	botInfo, err := bot.getLoginInfo()
	if err != nil {
		log.Fatal("获取Bot信息失败:", err)
		return err
	}

	bot.Info = botInfo
	bot.ResChan = make(map[string]chan *CqResponse)

	go bot.ResponseUpdater()

	bot.initialized = true

	log.Printf("Bot Info: %s(%d)", bot.Info.NickName, bot.Info.UserId)

	return nil
}

func (bot *Bot) IsInitialized() bool {
	return bot.initialized
}

// Just for API test
func (bot *Bot) getLoginInfo() (*BotInfo, error) {
	req := CqRequest{
		Action: "get_login_info",
	}

	reqJson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = bot.Client.ActConn.WriteMessage(websocket.TextMessage, reqJson)
	if err != nil {
		return nil, err
	}

	_, message, err := bot.Client.ActConn.ReadMessage()
	if err != nil {
		return nil, err
	}

	res := CqResponse{}
	err = json.Unmarshal(message, &res)
	if err != nil {
		return nil, err
	}

	if res.Status != "ok" {
		return nil, &ActionFailErr{res.Wording}
	}

	res.Json = gjson.ParseBytes(message)

	botInfo := BotInfo{
		UserId:   res.Json.Get("data.user_id").Int(),
		NickName: res.Json.Get("data.nickname").String(),
	}

	return &botInfo, nil
}

func (bot *Bot) ResponseUpdater() {
	//Handler
	go func(bot *Bot) {
		for {
			_, res, err := bot.Client.ActConn.ReadMessage()

			if err != nil {
				log.Fatal("与 cqHttp 的连接出错:", err)
				return
			}

			cqRes := &CqResponse{}
			err = json.Unmarshal(res, &cqRes)
			if err != nil {
				log.Fatal("cqHttp 事件获取失败:", err)
				return
			}

			cqRes.Json = gjson.ParseBytes(res)

			if bot.ResChan[cqRes.Echo] != nil {
				bot.ResChan[cqRes.Echo] <- cqRes
			}
		}
	}(bot)
}

// SendPrivateMessage 发送私聊信息
// message string 消息文本
// id int64 用户 ID
// autoEscape bool 消息内容是否作为纯文本发送 ( 即不解析 CQ 码 )
func (bot *Bot) SendPrivateMessage(message string, userId int64, autoEscape bool) error {
	req := CqRequest{
		Action: "send_private_msg",
		Params: map[string]interface{}{
			"message":     message,
			"user_id":     userId,
			"auto_escape": autoEscape,
		},
		Echo: uuid.NewV4().String(),
	}

	err := bot.doAction(&req)

	if err != nil {
		log.Println("发送消息失败:", err)
		return err
	}

	res := bot.getActionResult(req.Echo)
	if res.Status != "ok" {
		return &ActionFailErr{res.Wording}
	}

	return nil
}

// SendGroupMessage 发送群聊信息
// message string 消息文本
// id int64 群组 ID
func (bot *Bot) SendGroupMessage(message string, groupId int64, autoEscape bool) error {
	req := CqRequest{
		Action: "send_group_msg",
		Params: map[string]interface{}{
			"message":     message,
			"group_id":    groupId,
			"auto_escape": autoEscape,
		},
		Echo: uuid.NewV4().String(),
	}

	err := bot.doAction(&req)

	if err != nil {
		log.Println("发送消息失败:", err)
		return err
	}

	res := bot.getActionResult(req.Echo)
	if res.Status != "ok" {
		return &ActionFailErr{res.Wording}
	}

	return nil
}

// GetMessage 获取消息
// messageId int64 消息ID
func (bot *Bot) GetMessage(messageId int64) *Message {
	req := CqRequest{
		Action: "get_msg",
		Params: map[string]interface{}{
			"message_id": messageId,
		},
		Echo: uuid.NewV4().String(),
	}

	err := bot.doAction(&req)
	if err != nil {
		log.Println("获取消息失败:", err)
		return nil
	}

	res := bot.getActionResult(req.Echo)
	if res.Status != "ok" {
		return nil
	}

	msg := &Message{
		MessageType: res.Json.Get("data.message_type").String(),
		MessageID:   res.Json.Get("data.message_id").Int(),
		Sender: struct {
			UserId   int64  `json:"user_id"`
			NickName string `json:"nickname"`
			Card     string `json:"card"`
		}{
			UserId:   res.Json.Get("data.sender.user_id").Int(),
			NickName: res.Json.Get("data.sender.nickname").String(),
		},
		RawMessage: res.Json.Get("data.message").String(),

		Bot: bot,
	}

	if msg.IsGroupMessage() {
		msg.GroupId = res.Json.Get("data.group_id").Int()
	}

	return msg
}

func (bot *Bot) GetGroupList() [][]any {
	req := CqRequest{
		Action: "get_group_list",
		Echo:   uuid.NewV4().String(),
	}

	err := bot.doAction(&req)
	if err != nil {
		log.Println("获取群组列表失败:", err)
		return nil
	}

	res := bot.getActionResult(req.Echo)
	if res.Status != "ok" {
		return nil
	}

	groupList := make([][]any, 0)

	groups := res.Json.Get("data").Array()

	for _, group := range groups {
		groupId := group.Get("group_id").Int()
		groupName := group.Get("group_name").String()
		groupList = append(groupList, append(make([]any, 0), groupId, groupName))
	}
	return groupList
}
