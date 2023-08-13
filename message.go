package hareru_cq

import (
	"fmt"
	"regexp"
	"strconv"
)

type User struct {
	UserId   int64
	NickName string
}

type GroupMember struct {
	User            *User
	Card            string
	JoinTime        int64
	LastSentTime    int64
	Level           string
	Role            string
	CardChangeable  bool
	ShutUpTimeStamp int64
}

type Message struct {
	MessageType string `json:"message_type"`
	MessageID   int64  `json:"message_id"`
	Sender      *User
	RawMessage  string `json:"raw_message"`

	// if message_type is group, this field is available
	GroupId     int64 `json:"group_id"`
	GroupMember *GroupMember
	Bot         *Bot
}

func (msg *Message) IsGroupMessage() bool {
	return msg.MessageType == "group"
}

func (msg *Message) IsPrivateMessage() bool {
	return msg.MessageType == "private"
}

func (msg *Message) ReplyMessage(message string, explicit bool) error {
	if explicit {
		message = fmt.Sprintf("[CQ:reply,id=%d] %s", msg.MessageID, message)
	} else {
		message = fmt.Sprintf("[CQ:at,qq=%d] %s", msg.Sender.UserId, message)
	}

	if msg.IsGroupMessage() {
		err := msg.Bot.SendGroupMessage(message, msg.GroupId, explicit)
		if err != nil {
			return err
		}
	} else if msg.IsPrivateMessage() {
		err := msg.Bot.SendPrivateMessage(message, msg.Sender.UserId, explicit)
		if err != nil {
			return err
		}
	}

	return nil
}

func (msg *Message) GetRepliedMessage() *Message {
	re := regexp.MustCompile(`\[CQ:reply,id=(-?\d+)\]`)
	match := re.FindStringSubmatch(msg.RawMessage)

	if len(match) <= 1 {
		return nil
	}

	reMessageId, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return nil
	}

	reMessage, err := msg.Bot.GetMessage(reMessageId, true)
	if err != nil {
		return nil
	}

	return reMessage
}

func buildMessageByUpdate(update *Update) *Message {
	user := &User{
		UserId:   update.Event.Get("sender.user_id").Int(),
		NickName: update.Event.Get("sender.nickname").String(),
	}

	msg := Message{
		MessageType: update.Event.Get("message_type").String(),
		MessageID:   update.Event.Get("message_id").Int(),
		Sender:      user,
		RawMessage:  update.Event.Get("raw_message").String(),

		Bot: update.Bot,
	}

	if msg.IsGroupMessage() {
		msg.GroupId = update.Event.Get("group_id").Int()
		msg.GroupMember = &GroupMember{
			User: user,
			Card: update.Event.Get("sender.card").String(),

			Level: update.Event.Get("sender.level").String(),
			Role:  update.Event.Get("sender.card").String(),
		}
	}

	return &msg
}
