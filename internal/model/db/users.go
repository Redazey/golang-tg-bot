package db

// Работа с хранилищем информации о пользователях.

import (
	"context"

	types "tgssn/internal/model/bottypes"
	"tgssn/internal/utils/dbutils"
	"tgssn/pkg/errors"

	"github.com/jmoiron/sqlx"
)

// UserDataReportRecordDB - Тип, принимающий структуру записей о расходах.
type UserDataReportRecordDB struct {
	Category string `db:"name"`
	Sum      int64  `db:"sum"`
}

// UserStorage - Тип для хранилища информации о пользователях.
type UserStorage struct {
	db *sqlx.DB
}

// NewUserStorage - Инициализация хранилища информации о пользователях.
// db - *sqlx.DB - ссылка на подключение к БД.
func NewUserStorage(db *sqlx.DB) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

func (storage *UserStorage) GetUserOrdersCount(ctx context.Context, userID int64) (int64, error) {
	return 6969, nil
}

// InsertUser Добавление пользователя в базу данных.
func (storage *UserStorage) InsertUser(ctx context.Context, userID int64) error {
	// Запрос на добавление данных.
	const sqlString = `
		INSERT INTO users (tg_id, limits)
		VALUES ($1, 0)
		ON CONFLICT (tg_id) DO NOTHING;
	`

	// Выполнение запроса на добавление данных.
	if _, err := dbutils.Exec(ctx, storage.db, sqlString, userID); err != nil {
		return err
	}
	return nil
}

// CheckIfUserExist Проверка существования пользователя в базе данных.
func (storage *UserStorage) CheckIfUserExist(ctx context.Context, userID int64) (bool, error) {
	// Запрос на выборку пользователя.
	const sqlString = `SELECT COUNT(id) AS countusers FROM users WHERE tg_id = $1;`

	// Выполнение запроса на получение данных.
	cnt, err := dbutils.GetMap(ctx, storage.db, sqlString, userID)
	if err != nil {
		return false, err
	}
	// Приведение результата запроса к нужному типу.
	countusers, ok := cnt["countusers"].(int64)
	if !ok {
		return false, errors.New("Ошибка приведения типа результата запроса.")
	}
	if countusers == 0 {
		return false, nil
	}
	return true, nil
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

// InsertUserDataRecord Добавление записи о расходах пользователя (в транзакции с проверкой превышения лимита).
func (storage *UserStorage) InsertUserDataRecord(ctx context.Context, userID int64, rec types.UserDataRecord) (bool, error) {
	// Проверка существования пользователя в БД.
	_, err := storage.CheckIfUserExistAndAdd(ctx, userID)
	if err != nil {
		return false, err
	}

	// Проверка, что не превышен лимит расходов.
	limit, err := storage.GetUserLimit(ctx, userID)
	if err != nil {
		return false, err
	}
	if limit < rec.Sum {
		return false, nil
	}

	// Запуск транзакции.
	err = dbutils.RunTx(ctx, storage.db,
		// Функция, выполняемая внутри транзакции.
		// Если функция вернет ошибку, произойдет откат транзакции.
		func(tx *sqlx.Tx) error {
			err = insertUserDataRecordTx(ctx, tx, userID, rec)
			return err
		})

	return true, err
}

// GetUserLimit Получение бюджета пользователя.
func (storage *UserStorage) GetUserLimit(ctx context.Context, userID int64) (float64, error) {
	// Получение бюджета по пользователю.
	const sqlString = `SELECT limits FROM users WHERE tg_id = $1;`

	// Выполнение запроса на выборку данных (запись результата запроса в map).
	result, err := dbutils.GetMap(ctx, storage.db, sqlString, userID)
	if err != nil {
		return 0, errors.Wrap(err, "Get user limits error")
	}
	// Приведение результата запроса к нужному типу.
	limits, ok := result["limits"].(float64)
	if !ok {
		return 0, errors.New("Ошибка приведения типа результата запроса.")
	}
	return limits, nil
}

func (storage *UserStorage) AddUserLimit(ctx context.Context, userID int64, limits float64, userName string) error {
	// Проверка существования пользователя в БД.
	_, err := storage.CheckIfUserExistAndAdd(ctx, userID)
	if err != nil {
		return err
	}
	// Запрос на обновление данных.
	const sqlString = `UPDATE users 
					   SET limits = limits + $1 
					   WHERE tg_id = $2;`

	// Выполнение запроса на обновление данных.
	if _, err := dbutils.Exec(ctx, storage.db, sqlString, limits, userID); err != nil {
		return err
	}
	return nil
}

func checkIfUserRecordsExist(ctx context.Context, db sqlx.QueryerContext, userID int64) (bool, error) {
	// Запрос на проверку лимита.
	const sqlString = `
		SELECT COUNT(r.id) AS counter
		FROM users AS u
			INNER JOIN usermoneytransactions AS r
				ON r.user_id = u.id
		WHERE u.tg_id = $1 
		;`

	// Выполнение запроса на получение данных.
	cnt, err := dbutils.GetMap(ctx, db, sqlString, userID)
	if err != nil {
		return false, err
	}
	// Приведение результата запроса к нужному типу.
	counter, ok := cnt["counter"].(int64)
	if !ok {
		return false, errors.New("Ошибка приведения типа результата запроса.")
	}
	if counter == 0 {
		return false, nil
	}
	return true, nil
}

// insertUserDataRecordTx Функция добавления расхода, выполняемая внутри транзакции (tx).
func insertUserDataRecordTx(ctx context.Context, tx sqlx.ExtContext, userID int64, rec types.UserDataRecord) error {

	// Запрос на добаление записи с проверкой существования категории.
	const sqlString = `
		WITH rows AS (INSERT INTO usercategories (user_id, name)
		(SELECT id, :category_name FROM users WHERE users.tg_id = :tg_id)
		ON CONFLICT (user_id, lower(name)) DO NOTHING)
		INSERT INTO usermoneytransactions (user_id, category_id, sum)
		(SELECT u.id, c.id, :sum
		FROM usercategories AS c
		INNER JOIN users AS u ON c.user_id = u.id
		WHERE u.tg_id = :tg_id AND lower(c.name) = lower(:category_name))
		ON CONFLICT DO NOTHING;`

	// Именованные параметры запроса.
	args := map[string]any{
		"tg_id":         userID,
		"category_name": rec.Category,
		"sum":           rec.Sum,
	}

	// Запуск на выполнение запроса с именованными параметрами.
	if _, err := dbutils.NamedExec(ctx, tx, sqlString, args); err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return err
	}

	return nil
}
