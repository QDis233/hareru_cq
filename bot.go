package hareru_cq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"image"
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

	GroupOwnerRole  = "owner"
	GroupAdminRole  = "admin"
	GroupMemberRole = "member"
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
		log.Error("发送请求失败:", err)
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
	bot.initialized = false
}

func (bot *Bot) Init() error {
	if !bot.Client.IsInitialized() {
		err := bot.Client.Init()
		if err != nil {
			return err
		}
	}

	botInfo, err := bot.getBotInfo()
	if err != nil {
		log.Fatal("获取Bot信息失败:", err)
		return err
	}

	bot.Info = botInfo
	bot.ResChan = make(map[string]chan *CqResponse)

	go bot.ResponseUpdater()

	bot.initialized = true

	log.Infof("Bot Info: %s(%d)", bot.Info.NickName, bot.Info.UserId)

	return nil
}

func (bot *Bot) IsInitialized() bool {
	return bot.initialized
}

// getBotInfo 获取Bot信息
func (bot *Bot) getBotInfo() (*BotInfo, error) {
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
		return err
	}

	res := bot.getActionResult(req.Echo)
	if res.Status != "ok" {
		return &ActionFailErr{res.Wording}
	}

	return nil
}

// GetMessage 获取消息
// messageId int64 消息ID Not real_id
func (bot *Bot) GetMessage(messageId int64, furtherInfo bool) (*Message, error) {
	req := CqRequest{
		Action: "get_msg",
		Params: map[string]interface{}{
			"message_id": messageId,
		},
		Echo: uuid.NewV4().String(),
	}

	err := bot.doAction(&req)
	if err != nil {
		return nil, &ActionFailErr{Message: err.Error()}
	}

	res := bot.getActionResult(req.Echo)
	if res.Status != "ok" {
		return nil, &ActionFailErr{res.Wording}
	}

	msg := &Message{
		MessageType: res.Json.Get("data.message_type").String(),
		MessageID:   res.Json.Get("data.message_id").Int(),
		Sender: &User{
			UserId:   res.Json.Get("data.sender.user_id").Int(),
			NickName: res.Json.Get("data.sender.nickname").String(),
		},
		RawMessage: res.Json.Get("data.message").String(),

		Bot: bot,
	}

	if msg.IsGroupMessage() {
		groupId := res.Json.Get("data.group_id").Int()

		grpMember, err := bot.GetGroupMember(groupId, msg.Sender.UserId)
		if err != nil {
			return nil, err
		}

		msg.GroupId = groupId
		msg.GroupMember = grpMember
	}

	return msg, nil
}

func (bot *Bot) GetGroupList() [][]any {
	req := CqRequest{
		Action: "get_group_list",
		Echo:   uuid.NewV4().String(),
	}

	err := bot.doAction(&req)
	if err != nil {
		log.Error("获取群组列表失败:", err)
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

func (bot *Bot) GetAvatar(userId int64, size int) (image.Image, error) {
	url := fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=%d", userId, size)

	imageData, err := getHttpRes(url)
	if err != nil {
		log.Error("获取头像失败:", err)
		return nil, err
	}

	imageDataReader := bytes.NewReader(imageData)
	img, _, err := image.Decode(imageDataReader)
	if err != nil {
		log.Error("获取头像失败:", err)
		return nil, err
	}

	return img, nil
}

func (bot *Bot) GetGroupMember(groupId int64, userId int64) (*GroupMember, error) {
	req := CqRequest{
		Action: "get_group_member_info",
		Params: map[string]interface{}{
			"group_id": groupId,
			"user_id":  userId,
		},
		Echo: uuid.NewV4().String(),
	}

	err := bot.doAction(&req)
	if err != nil {
		return nil, &ActionFailErr{Message: "failed to get group member info " + err.Error()}
	}

	res := bot.getActionResult(req.Echo)
	if res.Status != "ok" {
		return nil, &ActionFailErr{Message: "failed to get group member info " + res.Wording}
	}

	grpMember := GroupMember{
		User: &User{
			UserId:   res.Json.Get("data.user_id").Int(),
			NickName: res.Json.Get("data.nickname").String(),
		},
		Card: res.Json.Get("data.card").String(),

		JoinTime:        res.Json.Get("data.join_time").Int(),
		LastSentTime:    res.Json.Get("data.last_sent_time").Int(),
		Level:           res.Json.Get("data.level").String(),
		Role:            res.Json.Get("data.role").String(),
		CardChangeable:  res.Json.Get("data.card_changeable").Bool(),
		ShutUpTimeStamp: res.Json.Get("data.shut_up_timestamp").Int(),
	}

	return &grpMember, nil
}
