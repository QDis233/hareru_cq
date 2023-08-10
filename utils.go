package hareru_cq

func buildMessage(update *Update) *Message {
	msg := Message{
		MessageType: update.Event.Get("message_type").String(),
		MessageID:   update.Event.Get("message_id").Int(),
		Sender: struct {
			UserId   int64  `json:"user_id"`
			NickName string `json:"nickname"`
			Card     string `json:"card"`
		}{
			UserId:   update.Event.Get("sender.user_id").Int(),
			NickName: update.Event.Get("sender.nickname").String(),
		},
		RawMessage: update.Event.Get("raw_message").String(),

		Bot: update.Bot,
	}

	if msg.IsGroupMessage() {
		msg.GroupId = update.Event.Get("group_id").Int()
	}

	return &msg
}
