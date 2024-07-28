package payment

import (
	"context"
	types "tgssn/internal/model/bottypes"
)

// UserDataStorage Интерфейс для работы с хранилищем данных.
type UserDataStorage interface {
	CheckIfUserExistAndAdd(ctx context.Context, userID int64) (bool, error)
	InsertUserDataRecord(ctx context.Context, userID int64, rec types.UserDataRecord) (bool, error)
	AddUserLimit(ctx context.Context, userID int64, limits float64) error
	GetUserLimit(ctx context.Context, userID int64) (float64, error)
	CheckIfUserRecordsExist(ctx context.Context, userID int64) (int64, error)
}

// Model Модель платёжки
type Model struct {
	ctx     context.Context
	storage UserDataStorage // Хранилище пользовательской информации.
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, storage UserDataStorage) *Model {
	return &Model{
		ctx:     ctx,
		storage: storage,
	}
}

func (s *Model) MockPay(sum int64) (bool, error) {
	return true, nil
}
