package hareru_cq

type Filter interface {
	Filter(args ...any) bool
}

type EventFilter struct {
	filterMap map[string]func(update *Update, eventType string) bool
}

func (f *EventFilter) messageFilter(update *Update, eventType string) bool {
	if update.Event.Type == "message" {
		if eventType == ReceiveMessageEvent {
			return true
		}

		msgType := update.Event.Get("message_type").String()

		if (msgType == "private" && eventType == PrivateMessageEvent) || (msgType == "group" && eventType == GroupMessageEvent) {
			return true
		}
	}

	return false
}

func (f *EventFilter) requestFilter(update *Update, eventType string) bool {
	if update.Event.Type == "request" {
		reqType := update.Event.Get("request_type").String()

		if (reqType == "friend" && eventType == FriendRequestEvent) || (reqType == "group" && eventType == GroupRequestEvent) {
			return true
		}
	}

	return false
}

func (f *EventFilter) noticeFilter(update *Update, eventType string) bool {
	if update.Event.Type == "notice" {
		noticeType := update.Event.Get("notice_type").String()

		// 偷个小懒
		if noticeType == eventType {
			return true
		}
	}

	return false
}

func (f *EventFilter) metaEventFilter(update *Update, eventType string) bool {
	if update.Event.Type == "meta_event" {
		metaEventType := update.Event.Get("meta_event_type").String()

		if metaEventType == eventType {
			return true
		}
	}

	return false
}

func (f *EventFilter) Filter(update *Update, eventType string) bool {
	f.filterMap = map[string]func(update *Update, eventType string) bool{
		ReceiveMessageEvent: f.messageFilter,
		PrivateMessageEvent: f.messageFilter,
		GroupMessageEvent:   f.messageFilter,

		FriendRecallEvent:  f.noticeFilter,
		GroupRecallEvent:   f.noticeFilter,
		GroupIncreaseEvent: f.noticeFilter,
		GroupDecreaseEvent: f.noticeFilter,
		GroupAdminEvent:    f.noticeFilter,
		GroupUploadEvent:   f.noticeFilter,
		GroupBanEvent:      f.noticeFilter,
		FriendAddEvent:     f.noticeFilter,
		GroupCardEvent:     f.noticeFilter,
		EssenceEvent:       f.noticeFilter,

		FriendRequestEvent: f.requestFilter,
		GroupRequestEvent:  f.requestFilter,

		LifecycleEvent: f.metaEventFilter,
		HeartbeatEvent: f.metaEventFilter,
	}

	return f.filterMap[eventType](update, eventType)
}

func NewEventFilter() EventFilter {
	return EventFilter{}
}
