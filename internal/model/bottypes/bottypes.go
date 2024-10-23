package bottypes

type Empty struct{}

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
