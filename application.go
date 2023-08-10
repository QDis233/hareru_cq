package hareru_cq

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Application struct {
	Name string

	Bot      *Bot
	Updater  *Updater
	Handlers []Handler

	initialized bool
	running     bool
}

func (app *Application) Init() error {
	if app.initialized {
		log.Fatal(app.Name, "Already initialized!")
		return &AlreadyInitializedErr{}
	}

	if !app.Bot.IsInitialized() {
		err := app.Bot.Init()
		if err != nil {
			return err
		}
	}

	if app.Updater == nil {
		log.Fatal("No Updater!")
		return &NotAvailableErr{
			"Updater not available",
		}
	}

	if !app.Updater.IsInitialized() {
		err := app.Updater.Init()
		if err != nil {
			return err
		}
	}

	app.initialized = true
	return nil
}

func (app *Application) AddHandler(handler Handler) {
	app.Handlers = append(app.Handlers, handler)
}

func (app *Application) RunPulling() {
	if !app.initialized {
		err := app.Init()
		if err != nil {
			panic(err)
		}
	}

	if app.running {
		log.Fatal(app.Name, " 已在运行中")
		return
	}

	app.exitHandler()
	app.processUpdate()

	return
}

func (app *Application) processUpdate() {
	app.running = true
	log.Println(app.Name, " 启动成功!")

	for {
		update := <-app.Updater.Updates

		for _, handler := range app.Handlers {
			check := handler.CheckUpdate(update)
			if check {
				handler.CollectArgs(update)
				go handler.HandleUpdate(update)
			}
		}
	}
}

func (app *Application) exitHandler() {
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-exitSignal
		log.Println("Stopping...")

		defer app.Bot.Stop()

		os.Exit(0)
	}()
}
