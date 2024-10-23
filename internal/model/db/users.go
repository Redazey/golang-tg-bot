package db

// Работа с хранилищем информации о пользователях.

import (
	"context"

	"tgseller/internal/model/bottypes"

	"gorm.io/gorm"
)

// UserStorage - Тип для хранилища информации о пользователях.
type UserStorage struct {
	db *gorm.DB
}

// NewUserStorage - Инициализация хранилища информации о пользователях.
// db - *sqlx.DB - ссылка на подключение к БД.
func NewUserStorage(db *gorm.DB) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

// InsertUser Добавление пользователя в базу данных.
func (storage *UserStorage) InsertUser(ctx context.Context, userID int64) error {
	tx := storage.db.Create(&bottypes.Users{ID: userID})
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// CheckIfUserExist Проверка существования пользователя в базе данных.
// false - не найдено | true - найдено
func (storage *UserStorage) CheckIfUserExist(ctx context.Context, userID int64) (bool, error) {
	var user bottypes.Users
	tx := storage.db.First(&user, userID)

	switch tx.Error {
	case nil:
		return false, tx.Error
	case gorm.ErrRecordNotFound:
		return false, nil
	default:
		return true, nil
	}
}

// CheckIfUserExistAndAdd Проверка существования пользователя в базе данных добавление, если не существует.
func (storage *UserStorage) CheckIfUserExistAndAdd(ctx context.Context, userID int64) (bool, error) {
	exist, err := storage.CheckIfUserExist(ctx, userID)
	if err != nil {
		return false, err
	}
	if !exist {
		// Добавление пользователя в базу, если не существует.
		err := storage.InsertUser(ctx, userID)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
