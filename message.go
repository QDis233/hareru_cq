package hareru_cq

import (
	"fmt"
)

type Message struct {
	MessageType string `json:"message_type"`
	MessageID   int64  `json:"message_id"`
	Sender      struct {
		UserId   int64  `json:"user_id"`
		NickName string `json:"nickname"`
		Card     string `json:"card"`
	}
	RawMessage string `json:"raw_message"`

	// if message_type is group, this field is available
	GroupId int64 `json:"group_id"`
	Bot     *Bot
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

func (msg *Message) NewMessage(messageId int64, bot *Bot) *Message {
	return bot.GetMessage(messageId)
}
