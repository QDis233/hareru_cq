package hareru_cq

import (
	"regexp"
	"strings"
)

// Handler Handler 接口
type Handler interface {
	CheckUpdate(update *Update) bool
	HandleUpdate(update *Update) any
	CollectArgs(update *Update) //Prepares additional arguments
}

// MessageHandler 消息处理器
type MessageHandler struct {
	Filter   Filter                      //消息过滤器
	Callback func(*Update, *Message) any //消息处理函数
}

func (h *MessageHandler) CheckUpdate(update *Update) bool {
	return h.Filter.Filter(update)
}

func (h *MessageHandler) HandleUpdate(update *Update) interface{} {
	message := buildMessage(update)
	return h.Callback(update, message)
}

func (h *MessageHandler) CollectArgs(update *Update) {
	return
}

// TextHandler 消息文本处理器
type TextHandler struct {
	MessagePattern string                      //消息匹配 正则表达式
	Callback       func(*Update, *Message) any //消息处理函数
}

func (h *TextHandler) CheckUpdate(update *Update) bool {
	filter := NewEventFilter()
	if filter.Filter(update, ReceiveMessageEvent) {
		re := regexp.MustCompile(h.MessagePattern)
		return re.MatchString(update.Event.Json.Get("message").String())
	}
	return false
}

func (h *TextHandler) HandleUpdate(update *Update) interface{} {
	message := buildMessage(update)
	return h.Callback(update, message)
}

// CollectArgs prepare args
func (h *TextHandler) CollectArgs(update *Update) {
	return
}

func NewTextHandler(pattern string, callback func(*Update, *Message) any) TextHandler {
	return TextHandler{
		MessagePattern: pattern,
		Callback:       callback,
	}
}

// CommandHandler 消息命令处理器
type CommandHandler struct {
	Command  string                      //命令 (写在!后面的部分)
	Callback func(*Update, *Message) any //消息处理函数

	args []string //消息参数
}

func (h *CommandHandler) CheckUpdate(update *Update) bool {
	filter := NewEventFilter()
	if filter.Filter(update, ReceiveMessageEvent) {
		return strings.HasPrefix(update.Event.Json.Get("message").String(), "!"+h.Command)
	}
	return false
}

func (h *CommandHandler) HandleUpdate(update *Update) interface{} {
	message := buildMessage(update)
	return h.Callback(update, message)
}

// CollectArgs prepare args
func (h *CommandHandler) CollectArgs(update *Update) {
	message := strings.TrimPrefix(update.Event.Get("message").String(), "!")
	h.args = strings.Split(message, " ")
}

func NewCommandHandler(command string, callback func(*Update, *Message) any) CommandHandler {
	return CommandHandler{
		Command:  command,
		Callback: callback,
	}
}
