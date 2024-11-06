package invoiceaccepter

import (
	"context"
	"tgseller/config"
	"tgseller/internal/model/bottypes"
	"tgseller/pkg/logger"
	"time"

	"go.uber.org/zap"
)

type UserDataStorage interface {
	GetAccessedUsers(ctx context.Context) ([]bottypes.Users, error)
	GetUnAccessedUsers(ctx context.Context) ([]bottypes.Users, error)
	ChangeUserAccess(ctx context.Context, userID int64, Status bool) error
	ChangeIsAddedStatus(ctx context.Context, userID int64, Status bool) error
}

type tgClient interface {
	AcceptInvoice(userID int64, chatID int64) (int64, error)
	DeleteUser(userID int64, chatID int64) (int64, error)
}

// Model Модель платёжки
type Model struct {
	ctx      context.Context
	cfg      *config.Enviroment
	storage  UserDataStorage // Хранилище пользовательской информации.
	tgClient tgClient        // Клиент.
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, cfg *config.Enviroment, storage UserDataStorage, tgClient tgClient) *Model {
	return &Model{
		ctx:      ctx,
		cfg:      cfg,
		storage:  storage,
		tgClient: tgClient,
	}
}

func (m *Model) Init() {
	go func() {
		for {
			m.CheckPendingRequests()
			m.CheckUnAccessedUsers()
			time.Sleep(time.Second * time.Duration(m.cfg.Dashboard))
		}
	}()
}

func (m *Model) CheckPendingRequests() {
	var users []bottypes.Users
	var err error
	// Получаем всех пользователей с Access = true
	if users, err = m.storage.GetAccessedUsers(m.ctx); err != nil {
		logger.Error("Error fetching users:", zap.Error(err))
		return
	}

	if users == nil {
		return
	}

	for _, user := range users {
		_, err := m.tgClient.AcceptInvoice(user.ID, m.cfg.PrivateChatID)
		if err != nil {
			logger.Error("Error while accepting invoice: ", zap.Error(err))
			continue
		}

		err = m.storage.ChangeIsAddedStatus(m.ctx, user.ID, true)
		if err != nil {
			logger.Error("Error while accepting invoice: ", zap.Error(err))
		}
	}
}

func (m *Model) CheckUnAccessedUsers() {
	var users []bottypes.Users
	var err error
	// Получаем всех пользователей с Access = true
	if users, err = m.storage.GetUnAccessedUsers(m.ctx); err != nil {
		logger.Error("Error fetching users:", zap.Error(err))
		return
	}

	for _, user := range users {
		err = m.storage.ChangeUserAccess(m.ctx, user.ID, false)
		if err != nil {
			logger.Error("Error approving request:", zap.Error(err))
			continue
		}

		_, err = m.tgClient.DeleteUser(user.ID, m.cfg.PrivateChatID)
		if err != nil {
			logger.Error("Error while deleting user: ", zap.Error(err))
			continue
		}

		err = m.storage.ChangeIsAddedStatus(m.ctx, user.ID, false)
		if err != nil {
			logger.Error("Error while approving invoice: ", zap.Error(err))
		}
	}
}
