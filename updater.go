package hareru_cq

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"

	"github.com/tidwall/gjson"
)

type Updater struct {
	Updates     chan *Update
	Bot         *Bot
	initialized bool
}

type Update struct {
	UpdateId int64
	Bot      *Bot
	Event    *Event
}

func (updater *Updater) Init() error {
	if updater.initialized {
		log.Error("Updater 已经初始化")
		return &AlreadyInitializedErr{}
	}

	if updater.Bot.initialized == false {
		log.Error("Bot 未初始化")
		return &NotAvailableErr{
			"Bot have not been initialized",
		}
	}

	go updater.startPull()
	updater.initialized = true
	return nil
}

func (updater *Updater) startPull() {
	for {
		_, message, err := updater.Bot.Client.EventConn.ReadMessage()
		if err != nil {
			log.Fatal("与 cqHttp 的连接出错:", err)
			return
		}

		event := &Event{}
		err = json.Unmarshal(message, &event)
		if err != nil {
			log.Fatal("cqHttp 事件获取失败:", err)
			return
		}

		event.Json = gjson.Parse(string(message))

		update := &Update{
			UpdateId: event.Time,
			Bot:      updater.Bot,
			Event:    event,
		}

		//log.Println("收到事件:", event.Json.Raw)

		updater.Updates <- update

	}

}

func (updater *Updater) IsInitialized() bool {
	return updater.initialized
}
