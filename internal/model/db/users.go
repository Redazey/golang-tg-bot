package db

// Работа с хранилищем информации о пользователях.

import (
	"context"
	"time"

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
	tx := storage.db.Create(&bottypes.Users{ID: userID, Access: false, Is_added: false})
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// InsertUser Добавление пользователя в базу данных.
func (storage *UserStorage) GetUserAccessStatus(ctx context.Context, user_id int64) (bool, error) {
	var user bottypes.Users
	tx := storage.db.Where("ID = ?", user_id).Select("Access").First(&user)
	if tx.Error != nil {
		tx.Rollback()
		return false, tx.Error
	}

	return user.Access, nil
}

// Поиск юзеров, которые получили доступ к каналу, но еще не состоят в нем
func (storage *UserStorage) GetAccessedUsers(ctx context.Context) ([]bottypes.Users, error) {
	var users []bottypes.Users
	tx := storage.db.Where("Access = ?", true).Where("Is_added = ?", false).Select("ID").Find(&users)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return users, nil
}

func (storage *UserStorage) GetUserAccessData(ctx context.Context, user_id int64) (time.Time, error) {
	var users bottypes.Users
	tx := storage.db.Where("ID = ?", user_id).Select("Accessed_at").First(&users)

	if tx.Error != nil {
		return time.Now(), tx.Error
	}

	return users.Accessed_at.AddDate(0, 1, 0), nil
}

func (storage *UserStorage) GetUnAccessedUsers(ctx context.Context) ([]bottypes.Users, error) {
	var users []bottypes.Users

	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	tx := storage.db.Where("Access = ?", true).Where("Is_added = ?", true).Where("Accessed_at < ?", oneMonthAgo).Select("ID").Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return users, nil
}

// InsertUser Добавление пользователя в базу данных.
func (storage *UserStorage) ChangeUserAccess(ctx context.Context, user_id int64, Status bool) error {
	if Status {
		access, err := storage.GetUserAccessStatus(ctx, user_id)
		if err != nil {
			return err
		}

		if access {
			tx := storage.db.Model(&bottypes.Users{}).
				Where("ID = ?", user_id).
				Update("Accessed_at", gorm.Expr("Accessed_at + INTERVAL '1 month'"))
			if tx.Error != nil {
				tx.Rollback()
				return tx.Error
			}
			return nil
		}
	}

	tx := storage.db.Model(&bottypes.Users{}).Where("ID = ?", user_id).Update("Access", Status)
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	tx = storage.db.Model(&bottypes.Users{}).Where("ID = ?", user_id).Update("Accessed_at", time.Now())
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return nil
}

// ChangeIsAddedStatus Установка статуса вступления в чат
func (storage *UserStorage) ChangeIsAddedStatus(ctx context.Context, user_id int64, Status bool) error {
	tx := storage.db.Model(&bottypes.Users{}).Where("ID = ?", user_id).Update("Is_added", Status)
	if tx.Error != nil {
		tx.Rollback()
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
	case gorm.ErrRecordNotFound:
		return false, nil
	}

	switch user {
	case bottypes.Users{}:
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

// InsertUserDataRecord Добавление записи о расходах пользователя
func (storage *UserStorage) InsertUserDataRecord(ctx context.Context, userID int64, ctgInfo bottypes.Records) (bool, error) {
	// Проверка существования пользователя в БД.
	_, err := storage.CheckIfUserExistAndAdd(ctx, userID)
	if err != nil {
		return false, err
	}

	tx := storage.db.Create(&ctgInfo)
	if tx.Error != nil {
		tx.Rollback()
		return false, tx.Error
	}

	return true, nil
}

func (storage *UserStorage) GetUserDataRecords(ctx context.Context) ([]bottypes.Records, error) {
	var records []bottypes.Records
	tx := storage.db.Where("status = ?", false).Find(&records)
	if tx.Error != nil {
		tx.Rollback()
		return nil, tx.Error
	}

	return records, nil
}

// InsertUser Добавление пользователя в базу данных.
func (storage *UserStorage) ChangeRecordStatus(ctx context.Context, record_id int64, Status bool) error {
	tx := storage.db.Model(&bottypes.Records{}).Where("ID = ?", record_id).Update("status", Status)
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return nil
}

// InsertUser Добавление пользователя в базу данных.
func (storage *UserStorage) DeleteUserRecord(ctx context.Context, record_id int64) error {
	var record bottypes.Records
	tx := storage.db.Delete(&record, record_id)
	if tx.Error != nil {
		tx.Rollback()
		return tx.Error
	}

	return nil
}
