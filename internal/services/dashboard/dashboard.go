package dashboard

import (
	"context"
	"fmt"
	"time"

	"tgssn/config"
	types "tgssn/internal/model/bottypes"
	consts "tgssn/internal/model/messages"
	"tgssn/pkg/logger"

	"go.uber.org/zap"
)

// Область "Внешний интерфейс": начало.

// MessageSender Интерфейс для работы с сообщениями.
type MessageSender interface {
	SendMessage(text string, userID int64) (int, error)
	ShowInlineButtons(text string, buttons []types.TgRowButtons, userID int64) (int, error)
	DeleteMsg(userID int64, msgID int)
}

// UserDataStorage Интерфейс для работы с хранилищем данных.
type UserDataStorage interface {
	GetAllWorkers(ctx context.Context) ([]int64, error)
	CountWorkersStatistic(ctx context.Context, workerID int64) (int64, int64, error)
	GetWorkerName(ctx context.Context, userID int64) (string, error)
}

type Model struct {
	ctx         context.Context
	tgClient    MessageSender   // Клиент.
	storage     UserDataStorage // Хранилище пользовательской информации.
	cfg         *config.Enviroment
	dashboardID int
	workersIDS  []int64
}

func New(ctx context.Context, tgClient MessageSender, storage UserDataStorage, cfg *config.Enviroment) *Model {
	return &Model{
		ctx:         ctx,
		tgClient:    tgClient,
		storage:     storage,
		cfg:         cfg,
		dashboardID: 0,
		workersIDS:  []int64{},
	}
}

func (s *Model) Dashboard() error {
	ctx := s.ctx
	var err error

	s.workersIDS, err = s.storage.GetAllWorkers(ctx)
	if err != nil {
		return err
	}

	var TxtWorkersStats = ""

	for _, worker := range s.workersIDS {
		goods, fails, err := s.storage.CountWorkersStatistic(ctx, worker)
		if err != nil {
			logger.Error("Ошибка при инициализации дэшборда: ", zap.Error(err))
			continue
		}

		name, err := s.storage.GetWorkerName(ctx, worker)
		if err != nil {
			logger.Error("Ошибка при инициализации дэшборда: ", zap.Error(err))
			continue
		}

		TxtWorkersStats += fmt.Sprintf(consts.TxtDashboardStats, name, goods, fails)

	}

	s.dashboardID, err = s.tgClient.SendMessage(
		fmt.Sprintf(consts.TxtDashboard, TxtWorkersStats),
		consts.WorkersChatID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Model) Init() {
	go func() {
		for {
			err := s.Dashboard()
			if err != nil {
				logger.Error("Failed to get updates", zap.Error(err))
				time.Sleep(time.Minute * time.Duration(s.cfg.Dashboard))

				continue
			}

			time.Sleep(time.Minute * time.Duration(s.cfg.Dashboard))

			s.tgClient.DeleteMsg(consts.WorkersChatID, s.dashboardID)
		}
	}()
}
