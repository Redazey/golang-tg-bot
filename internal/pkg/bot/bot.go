package bot

import (
	"context"
	"tgseller/config"
	"tgseller/internal/clients/tg"
	userStorage "tgseller/internal/model/db"
	"tgseller/internal/model/messages"
	"tgseller/internal/services/dashboard"
	"tgseller/internal/services/payment"
	"tgseller/pkg/cache"
	"tgseller/pkg/db"
	"tgseller/pkg/logger"
)

type App struct {
	tgClient  *tg.Client
	storage   *userStorage.UserStorage
	msgModel  *messages.Model
	payment   *payment.Model
	dashboard *dashboard.Model
}

func Init() (*App, error) {
	a := &App{}

	ctx := context.Background()

	cfg, err := config.NewEnv()
	if err != nil {
		return nil, err
	}

	logger.Init(cfg.LoggerLevel, "")

	err = db.Init(cfg.DB.DBUser, cfg.DB.DBPassword, cfg.DB.DBHost, cfg.DB.DBName)
	if err != nil {
		return nil, err
	}

	err = cache.Init(cfg.Redis.RedisAddr+":"+cfg.Redis.RedisPort, cfg.Redis.RedisPassword, 0, cfg.Cache.EXTime)
	if err != nil {
		return nil, err
	}

	a.tgClient, err = tg.New(cfg.TgToken, tg.HandlerFunc(tg.ProcessingMessages))
	if err != nil {
		return nil, err
	}

	a.storage = userStorage.NewUserStorage(db.GetDBConn())
	a.dashboard = dashboard.New(ctx, a.tgClient, a.storage, cfg)
	a.payment = payment.New(ctx, a.storage, a.tgClient, cfg.PaymentToken)
	a.msgModel = messages.New(ctx, a.tgClient, a.storage, a.payment, cfg)

	a.dashboard.Init()
	a.payment.Init()

	return a, nil
}

func (a *App) Run() error {
	logger.Info("Запуск бота")

	a.tgClient.ListenUpdates(a.msgModel)

	return nil
}
