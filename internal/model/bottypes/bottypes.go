package bottypes

import (
	"time"
)

type Empty struct{}

// Множество уникальных категорий покупок пользователя
type UserCategorySet map[string]Empty

type CtgInfo struct {
	ID          int64   `db:"id"`
	Name        string  `db:"name"`
	Price       float64 `db:"price"`
	Short       string  `db:"short_name"`
	Description string  `db:"description"`
	DataFormat  string  `db:"data_format"`
}

// Тип для записей о тратах.
type UserDataRecord struct {
	RecordID   int64
	UserID     int64
	CategoryID int64
	Period     time.Time
}

// Тип для записей о тратах.
type UserRefillRecord struct {
	RecordID  int64
	UserID    int64
	Status    string
	InvoiceID int64
	Amount    float64
	Period    time.Time
}

// Типы для описания состава кнопок телеграм сообщения.
// Кнопка сообщения.
type TgInlineButton struct {
	DisplayName string
	Value       string
	URL         string
}

// Строка с кнопками сообщения.
type TgRowButtons []TgInlineButton

// Типы для описания состава кнопок телеграм сообщения.
// Кнопка сообщения.
type TgKeyboardButton struct {
	Text string
}

// Строка с кнопками сообщения.
type TgKbRowButtons []TgKeyboardButton
