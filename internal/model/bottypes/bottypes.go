package bottypes

import (
	"time"
)

type Empty struct{}

// Множество уникальных категорий покупок пользователя
type UserCategorySet map[string]Empty

// Тип для записей о тратах.
type UserDataRecord struct {
	RecordID   int64
	UserID     int64
	CategoryID int64
	Period     time.Time
}

// Типы для описания состава кнопок телеграм сообщения.
// Кнопка сообщения.
type TgInlineButton struct {
	DisplayName string
	Value       string
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

// Тип для хранения курса валюты в формате "USD" = 0.01659657
type ExchangeRate map[string]float64
