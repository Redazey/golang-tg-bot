package db

// Работа с хранилищем информации о пользователях.

import (
	"context"
	"fmt"
	"time"

	types "tgssn/internal/model/bottypes"
	"tgssn/internal/utils/dbutils"
	"tgssn/pkg/errors"
	"tgssn/pkg/logger"

	"github.com/jmoiron/sqlx"
)

// UserDataReportRecordDB - Тип, принимающий структуру записей о расходах.
type UserDataReportRecordDB struct {
	Category string    `db:"name"`
	Sum      int64     `db:"sum"`
	Time     time.Time `db:"timestamp"`
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

func (storage *UserStorage) GetCategoryInfo(ctx context.Context, CtgID int64) (map[string]any, error) {
	// Запрос на добавление данных.
	const sqlString = `
		SELECT *
		FROM usercategories
		WHERE id = $1;
	`

	// Выполнение запроса на добавление данных.
	CtgInfo, err := dbutils.GetMap(ctx, storage.db, sqlString, CtgID)
	if err != nil {
		return nil, err
	}

	return CtgInfo, nil
}

func (storage *UserStorage) GetCtgInfoFromName(ctx context.Context, name string) (map[string]any, error) {
	// Запрос на добавление данных.
	const sqlString = `
		SELECT *
		FROM usercategories
		WHERE name = $1;
	`

	// Выполнение запроса на добавление данных.
	CtgInfo, err := dbutils.GetMap(ctx, storage.db, sqlString, name)
	if err != nil {
		return nil, err
	}

	return CtgInfo, nil
}

// InsertUser Добавление пользователя в базу данных.
func (storage *UserStorage) InsertUser(ctx context.Context, userID int64) error {
	// Запрос на добавление данных.
	const sqlString = `
		INSERT INTO users (tg_id)
		VALUES ($1)
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
	const sqlString = `SELECT COUNT(tg_id) AS countusers FROM users WHERE tg_id = $1;`

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

	ctgInfo, err := storage.GetCategoryInfo(ctx, rec.CategoryID)
	if err != nil {
		return true, err
	}

	ctgPrice, ok := ctgInfo["price"].(float64)
	if !ok {
		return true, errors.New("Ошибка при конвертации типов")
	}

	if limit < ctgPrice {
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

func (storage *UserStorage) AddUserLimit(ctx context.Context, userID int64, limits float64) error {
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

func (storage *UserStorage) CheckIfUserRecordsExist(ctx context.Context, userID int64) (int64, error) {
	// Запрос на проверку лимита.
	const sqlString = `
		SELECT COUNT(r.id) AS counter
		FROM users AS u
			INNER JOIN usermoneytransactions AS r
				ON r.tg_id = u.tg_id
		WHERE u.tg_id = $1 
		;`

	// Выполнение запроса на получение данных.
	cnt, err := dbutils.GetMap(ctx, storage.db, sqlString, userID)
	if err != nil {
		return 0, err
	}
	// Приведение результата запроса к нужному типу.
	counter, ok := cnt["counter"].(int64)
	if !ok {
		return 0, errors.New("Ошибка приведения типа результата запроса.")
	}
	return counter, nil
}

func (storage *UserStorage) GetUserOrders(ctx context.Context, userID int64) ([]types.UserDataRecord, error) {
	// Запрос на добаление записи с проверкой существования категории.
	const sqlString = `
		SELECT id, category_id, timestamp
		FROM usermoneytransactions
		WHERE tg_id = $1;`

	rows, err := storage.db.Query(sqlString, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []types.UserDataRecord

	for rows.Next() {
		var order struct {
			ID         int64
			CategoryID int64
			Timestamp  time.Time
		}

		// Считаем, что поле 'timestamp' имеет тип time.Time в базе данных
		if err := rows.Scan(&order.ID, &order.CategoryID, &order.Timestamp); err != nil {
			return nil, err
		}

		record := types.UserDataRecord{
			RecordID:   order.ID,
			UserID:     userID,
			CategoryID: order.CategoryID,
			Period:     order.Timestamp,
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

// insertUserDataRecordTx Функция добавления расхода, выполняемая внутри транзакции (tx).
func insertUserDataRecordTx(ctx context.Context, tx sqlx.ExtContext, userID int64, rec types.UserDataRecord) error {

	// Запрос на добаление записи с проверкой существования категории.
	const sqlString = `
		INSERT INTO usermoneytransactions (tg_id, category_id)
		VALUES (:tg_id, :category_id)`

	// Именованные параметры запроса.
	args := map[string]any{
		"tg_id":       userID,
		"category_id": rec.CategoryID,
	}

	// Запуск на выполнение запроса с именованными параметрами.
	if _, err := dbutils.NamedExec(ctx, tx, sqlString, args); err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return err
	}

	return nil
}

// Раздел функции для работников

// CheckIfWorkerExist Проверка существования пользователя в базе данных.
func (storage *UserStorage) CheckIfWorkerExist(ctx context.Context, userID int64) (bool, error) {
	// Запрос на выборку пользователя.
	const sqlString = `SELECT COUNT(tg_id) AS countworkers FROM workers WHERE tg_id = $1;`

	// Выполнение запроса на получение данных.
	cnt, err := dbutils.GetMap(ctx, storage.db, sqlString, userID)
	if err != nil {
		return false, err
	}
	// Приведение результата запроса к нужному типу.
	countworkers, ok := cnt["countworkers"].(int64)
	if !ok {
		return false, errors.New("Ошибка приведения типа результата запроса.")
	}
	if countworkers == 0 {
		return false, nil
	}
	return true, nil
}

// InsertWorker Добавление работника в базу данных.
func insertWorker(ctx context.Context, db *sqlx.DB, userID int64, name string) error {
	// Запрос на добавление данных.
	const sqlString = `
		INSERT INTO workers (tg_id, name)
		VALUES ($1, $2)
		ON CONFLICT (tg_id) DO NOTHING;`

	// Выполнение запроса на добавление данных.
	if _, err := dbutils.Exec(ctx, db, sqlString, userID, name); err != nil {
		return err
	}
	return nil
}

// CheckIfWorkerExistAndAdd Проверка существования пользователя в базе данных добавление, если не существует.
func (storage *UserStorage) CheckIfWorkerExistAndAdd(ctx context.Context, userID int64, name string) (bool, error) {
	exist, err := storage.CheckIfWorkerExist(ctx, userID)
	if err != nil {
		return false, err
	}
	if !exist {
		// Добавление пользователя в базу, если не существует.
		err := insertWorker(ctx, storage.db, userID, name)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (storage *UserStorage) GetAllWorkers(ctx context.Context) ([]int64, error) {
	// Запрос на добавление данных.
	const sqlString = `
		SELECT tg_id
		FROM workers;`

	// Именованные параметры запроса.
	res := []int64{}

	if err := dbutils.Select(ctx, storage.db, &res, sqlString); err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return nil, err
	}
	return res, nil
}

func (storage *UserStorage) getWorkerStatus(ctx context.Context, userID int64) (bool, error) {
	const sqlString = `
		SELECT status
		FROM workers
		WHERE tg_id = $1`

	res, err := dbutils.GetMap(ctx, storage.db, sqlString, userID)
	if err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return false, err
	}
	currStatus, ok := res["status"].(bool)
	if !ok {
		return false, err
	}

	if currStatus {
		return true, nil
	}

	return false, nil
}

func (storage *UserStorage) GetWorkerName(ctx context.Context, userID int64) (string, error) {
	const sqlString = `
		SELECT name
		FROM workers
		WHERE tg_id = $1`

	res, err := dbutils.GetMap(ctx, storage.db, sqlString, userID)
	if err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return "", err
	}

	name, ok := res["name"].(string)
	if !ok {
		return "", err
	}

	return name, nil
}

// возвращает был ли обновлен статус работника
func (storage *UserStorage) ChangeWorkerStatus(ctx context.Context, userID int64, status bool) (bool, error) {
	currStatus, err := storage.getWorkerStatus(ctx, userID)
	if err != nil {
		return false, err
	}

	if currStatus == status {
		return false, nil
	}

	const sqlString = `
		UPDATE workers
		SET status = $1
		WHERE tg_id = $2`

	if _, err := dbutils.Exec(ctx, storage.db, sqlString, status, userID); err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return false, err
	}

	return true, nil
}

// Функционал, относящийся к работе с тикетами
// Возвращает true, в случае если удалось создать тикет, false во всех остальных
func (storage *UserStorage) CreateTicket(ctx context.Context, workerID int64, buyerID int64, categoryID int64) (bool, error) {
	if succsessful, err := storage.ChangeWorkerStatus(ctx, workerID, true); !succsessful || err != nil {
		return false, err
	}
	// Запрос на добавление данных.
	const sqlString = `
		INSERT INTO tickets (worker_tg_id, buyer_tg_id, category_id)
		VALUES ($1, $2, $3);`

	if _, err := dbutils.Exec(ctx, storage.db, sqlString, workerID, buyerID, categoryID); err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return false, err
	}

	return true, nil
}

func (storage *UserStorage) UpdateTicketStatus(ctx context.Context, workerID int64, status string) error {
	// Запрос на добавление данных.
	const sqlString = `
		UPDATE tickets 
		SET status = $1
		WHERE worker_tg_id = $2 AND status = 'in_progress';`

	if _, err := dbutils.Exec(ctx, storage.db, sqlString, status, workerID); err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return err
	}
	return nil
}

// возвращает datamap с полями
// id, worker_tg_id, buyer_tg_id, category_id, status, timestamp
func (storage *UserStorage) GetTicketInfo(ctx context.Context, workerID int64) (map[string]any, error) {
	// Запрос на добавление данных.
	const sqlString = `
		SELECT *
		FROM tickets
		WHERE worker_tg_id = $1 AND status = 'in_progress';`

	dataMap, err := dbutils.GetMap(ctx, storage.db, sqlString, workerID)
	if err != nil {
		// Ошибка выполнения запроса (вызовет откат транзакции).
		return nil, err
	}

	return dataMap, nil
}

// возвращает статистику по тикетам определенного воркера в виде goods - bads - error
func (storage *UserStorage) CountWorkersStatistic(ctx context.Context, workerID int64) (int64, int64, error) {
	// Запрос на добавление данных.
	const sqlString = `
	SELECT
		COUNT(CASE WHEN status = 'good' THEN 1 END) AS goods,
		COUNT(CASE WHEN status = 'bad' THEN 1 END) AS bads
	FROM tickets
	WHERE worker_tg_id = $1
	AND DATE(timestamp) >= CURRENT_DATE;`

	// Выполнение запроса на получение данных.
	stats, err := dbutils.GetMap(ctx, storage.db, sqlString, workerID)
	if err != nil {
		return 0, 0, err
	}
	// Приведение результата запроса к нужному типу.
	goods, ok := stats["goods"].(int64)
	if !ok {
		return 0, 0, errors.New("Ошибка приведения типа результата запроса.")
	}

	bads, ok := stats["bads"].(int64)
	if !ok {
		logStr := fmt.Sprintf("Строка, с которой произошла ошибка: %v", stats["bads"])
		logger.Info(logStr)
		return 0, 0, errors.New("Ошибка приведения типа результата запроса.")
	}
	return goods, bads, nil
}
