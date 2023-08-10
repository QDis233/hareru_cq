package hareru_cq

type ApplicationBuilder struct {
	Name    string
	Client  *Client
	Bot     *Bot
	Updater *Updater
}

func NewApplicationBuilder() *ApplicationBuilder {
	return &ApplicationBuilder{}
}

func (builder *ApplicationBuilder) Build(appName string, apiUrl string) *Application {
	builder.Name = appName
	builder.Client = NewClient(apiUrl, "")
	builder.Bot = &Bot{
		Client: builder.Client,
	}
	builder.Updater = &Updater{
		Updates: make(chan *Update, 100),
		Bot:     builder.Bot,
	}

	app := Application{
		Name:     builder.Name,
		Bot:      builder.Bot,
		Updater:  builder.Updater,
		Handlers: make([]Handler, 0),
	}

	err := app.Init()
	if err != nil {
		return nil
	}

	return &app
}
